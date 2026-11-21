package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pedrobarros/vitaltag-api/internal/model"
)

type PerfilRepo struct {
	db *pgxpool.Pool
}

func NewPerfilRepo(db *pgxpool.Pool) *PerfilRepo {
	return &PerfilRepo{db: db}
}

// BuscarPorUsuario returns the user's complete profile.
func (r *PerfilRepo) BuscarPorUsuario(ctx context.Context, usuarioID uuid.UUID) (*model.PerfilCompleto, error) {
	p := &model.PerfilCompleto{}
	err := r.db.QueryRow(ctx,
		`SELECT id, usuario_id, nome, sobrenome, sexo, tipo_sanguineo,
		        slug_publico, senha_publica_hash, ativo, bloqueado_ate,
		        criado_em, atualizado_em
		 FROM perfil WHERE usuario_id = $1`,
		usuarioID,
	).Scan(
		&p.ID, &p.UsuarioID, &p.Nome, &p.Sobrenome,
		&p.Sexo, &p.TipoSanguineo, &p.SlugPublico,
		&p.SenhaPublicaHash, &p.Ativo, &p.BloqueadoAte,
		&p.CriadoEm, &p.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNaoEncontrado
		}
		return nil, fmt.Errorf("buscar perfil: %w", err)
	}

	if err := r.carregarDadosClinico(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// BuscarPorSlug returns a profile by public slug (for emergency access).
func (r *PerfilRepo) BuscarPorSlug(ctx context.Context, slug uuid.UUID) (*model.PerfilCompleto, error) {
	p := &model.PerfilCompleto{}
	err := r.db.QueryRow(ctx,
		`SELECT id, usuario_id, nome, sobrenome, sexo, tipo_sanguineo,
		        slug_publico, senha_publica_hash, ativo, bloqueado_ate,
		        criado_em, atualizado_em
		 FROM perfil WHERE slug_publico = $1`,
		slug,
	).Scan(
		&p.ID, &p.UsuarioID, &p.Nome, &p.Sobrenome,
		&p.Sexo, &p.TipoSanguineo, &p.SlugPublico,
		&p.SenhaPublicaHash, &p.Ativo, &p.BloqueadoAte,
		&p.CriadoEm, &p.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNaoEncontrado
		}
		return nil, fmt.Errorf("buscar perfil por slug: %w", err)
	}

	if err := r.carregarDadosClinico(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// SalvarOuAtualizar creates or replaces the profile and all clinical data in a transaction.
func (r *PerfilRepo) SalvarOuAtualizar(ctx context.Context, usuarioID uuid.UUID, req *model.PerfilRequest, senhaPublicaHash string) (*model.PerfilCompleto, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert the main profile
	p := &model.PerfilCompleto{}
	err = tx.QueryRow(ctx,
		`INSERT INTO perfil (usuario_id, nome, sobrenome, sexo, tipo_sanguineo, senha_publica_hash)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (usuario_id) DO UPDATE SET
		   nome = EXCLUDED.nome,
		   sobrenome = EXCLUDED.sobrenome,
		   sexo = EXCLUDED.sexo,
		   tipo_sanguineo = EXCLUDED.tipo_sanguineo,
		   senha_publica_hash = EXCLUDED.senha_publica_hash,
		   atualizado_em = now()
		 RETURNING id, usuario_id, nome, sobrenome, sexo, tipo_sanguineo,
		           slug_publico, senha_publica_hash, ativo, bloqueado_ate,
		           criado_em, atualizado_em`,
		usuarioID, req.Nome, req.Sobrenome, req.Sexo, req.TipoSanguineo, senhaPublicaHash,
	).Scan(
		&p.ID, &p.UsuarioID, &p.Nome, &p.Sobrenome,
		&p.Sexo, &p.TipoSanguineo, &p.SlugPublico,
		&p.SenhaPublicaHash, &p.Ativo, &p.BloqueadoAte,
		&p.CriadoEm, &p.AtualizadoEm,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert perfil: %w", err)
	}

	// Delete and recreate clinical data (full replacement — RF12)
	tables := []string{"contato_emergencia", "alergia", "medicamento", "doenca", "cirurgia"}
	for _, t := range tables {
		if _, err := tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE perfil_id = $1", t), p.ID); err != nil {
			return nil, fmt.Errorf("limpar %s: %w", t, err)
		}
	}

	// Emergency contacts
	for _, c := range req.Contatos {
		if _, err := tx.Exec(ctx,
			`INSERT INTO contato_emergencia (perfil_id, nome, parentesco, telefone, principal)
			 VALUES ($1,$2,$3,$4,$5)`,
			p.ID, c.Nome, c.Parentesco, c.Telefone, c.Principal,
		); err != nil {
			return nil, fmt.Errorf("inserir contato: %w", err)
		}
	}

	// Allergies
	for _, a := range req.Alergias {
		if _, err := tx.Exec(ctx,
			`INSERT INTO alergia (perfil_id, descricao, gravidade) VALUES ($1,$2,$3)`,
			p.ID, a.Descricao, nullIfEmpty(a.Gravidade),
		); err != nil {
			return nil, fmt.Errorf("inserir alergia: %w", err)
		}
	}

	// Medications
	for _, m := range req.Medicamentos {
		if _, err := tx.Exec(ctx,
			`INSERT INTO medicamento (perfil_id, nome, dosagem, frequencia) VALUES ($1,$2,$3,$4)`,
			p.ID, m.Nome, nullIfEmpty(m.Dosagem), nullIfEmpty(m.Frequencia),
		); err != nil {
			return nil, fmt.Errorf("inserir medicamento: %w", err)
		}
	}

	// Conditions
	for _, d := range req.Doencas {
		if _, err := tx.Exec(ctx,
			`INSERT INTO doenca (perfil_id, descricao, observacao) VALUES ($1,$2,$3)`,
			p.ID, d.Descricao, nullIfEmpty(d.Observacao),
		); err != nil {
			return nil, fmt.Errorf("inserir doenca: %w", err)
		}
	}

	// Surgeries
	for _, c := range req.Cirurgias {
		if _, err := tx.Exec(ctx,
			`INSERT INTO cirurgia (perfil_id, descricao, ano_realizacao) VALUES ($1,$2,$3)`,
			p.ID, c.Descricao, c.AnoRealizacao,
		); err != nil {
			return nil, fmt.Errorf("inserir cirurgia: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Reload clinical data for the response
	if err := r.carregarDadosClinico(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// RegistrarTentativa records a public-access attempt (RF09 / LGPD).
func (r *PerfilRepo) RegistrarTentativa(ctx context.Context, perfilID uuid.UUID, ip string, sucesso bool) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO tentativa_acesso (perfil_id, ip, sucesso) VALUES ($1, $2::inet, $3)`,
		perfilID, ip, sucesso,
	)
	return err
}

// ContarFalhasRecentes counts failed attempts in the last 15 minutes (RF09).
func (r *PerfilRepo) ContarFalhasRecentes(ctx context.Context, perfilID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tentativa_acesso
		 WHERE perfil_id = $1 AND sucesso = false AND data_hora > now() - interval '15 minutes'`,
		perfilID,
	).Scan(&count)
	return count, err
}

// Bloquear sets bloqueado_ate for the profile (RF09).
func (r *PerfilRepo) Bloquear(ctx context.Context, perfilID uuid.UUID, ate time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE perfil SET bloqueado_ate = $1 WHERE id = $2`,
		ate, perfilID,
	)
	return err
}

// ---------- internal helpers --------------------------------------------------

func (r *PerfilRepo) carregarDadosClinico(ctx context.Context, p *model.PerfilCompleto) error {
	// Emergency contacts
	rows, err := r.db.Query(ctx,
		`SELECT id, perfil_id, nome, COALESCE(parentesco,''), telefone, principal
		 FROM contato_emergencia WHERE perfil_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("buscar contatos: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c model.ContatoEmergencia
		if err := rows.Scan(&c.ID, &c.PerfilID, &c.Nome, &c.Parentesco, &c.Telefone, &c.Principal); err != nil {
			return err
		}
		p.Contatos = append(p.Contatos, c)
	}

	// Allergies
	aRows, err := r.db.Query(ctx,
		`SELECT id, perfil_id, descricao, COALESCE(gravidade,'') FROM alergia WHERE perfil_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("buscar alergias: %w", err)
	}
	defer aRows.Close()
	for aRows.Next() {
		var a model.Alergia
		if err := aRows.Scan(&a.ID, &a.PerfilID, &a.Descricao, &a.Gravidade); err != nil {
			return err
		}
		p.Alergias = append(p.Alergias, a)
	}

	// Medications
	mRows, err := r.db.Query(ctx,
		`SELECT id, perfil_id, nome, COALESCE(dosagem,''), COALESCE(frequencia,'')
		 FROM medicamento WHERE perfil_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("buscar medicamentos: %w", err)
	}
	defer mRows.Close()
	for mRows.Next() {
		var m model.Medicamento
		if err := mRows.Scan(&m.ID, &m.PerfilID, &m.Nome, &m.Dosagem, &m.Frequencia); err != nil {
			return err
		}
		p.Medicamentos = append(p.Medicamentos, m)
	}

	// Conditions
	dRows, err := r.db.Query(ctx,
		`SELECT id, perfil_id, descricao, COALESCE(observacao,'') FROM doenca WHERE perfil_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("buscar doencas: %w", err)
	}
	defer dRows.Close()
	for dRows.Next() {
		var d model.Doenca
		if err := dRows.Scan(&d.ID, &d.PerfilID, &d.Descricao, &d.Observacao); err != nil {
			return err
		}
		p.Doencas = append(p.Doencas, d)
	}

	// Surgeries
	cRows, err := r.db.Query(ctx,
		`SELECT id, perfil_id, descricao, ano_realizacao FROM cirurgia WHERE perfil_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("buscar cirurgias: %w", err)
	}
	defer cRows.Close()
	for cRows.Next() {
		var c model.Cirurgia
		if err := cRows.Scan(&c.ID, &c.PerfilID, &c.Descricao, &c.AnoRealizacao); err != nil {
			return err
		}
		p.Cirurgias = append(p.Cirurgias, c)
	}

	// Ensure slices are not null in JSON
	if p.Contatos == nil {
		p.Contatos = []model.ContatoEmergencia{}
	}
	if p.Alergias == nil {
		p.Alergias = []model.Alergia{}
	}
	if p.Medicamentos == nil {
		p.Medicamentos = []model.Medicamento{}
	}
	if p.Doencas == nil {
		p.Doencas = []model.Doenca{}
	}
	if p.Cirurgias == nil {
		p.Cirurgias = []model.Cirurgia{}
	}

	return nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
