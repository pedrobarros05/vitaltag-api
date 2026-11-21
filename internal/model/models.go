package model

import (
	"time"

	"github.com/google/uuid"
)

// ---------- User --------------------------------------------------------------

type Usuario struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	SenhaHash     string    `json:"-"`
	Consentimento bool      `json:"consentimento"`
	CriadoEm      time.Time `json:"criado_em"`
	AtualizadoEm  time.Time `json:"atualizado_em"`
}

// ---------- Profile -----------------------------------------------------------

type SexoTipo string
type TipoSanguineoTipo string

const (
	SexoFeminino     SexoTipo = "feminino"
	SexoMasculino    SexoTipo = "masculino"
	SexoOutro        SexoTipo = "outro"
	SexoNaoInformado SexoTipo = "nao_informado"
)

type Perfil struct {
	ID               uuid.UUID         `json:"id"`
	UsuarioID        uuid.UUID         `json:"usuario_id"`
	Nome             string            `json:"nome"`
	Sobrenome        string            `json:"sobrenome"`
	Sexo             SexoTipo          `json:"sexo"`
	TipoSanguineo    TipoSanguineoTipo `json:"tipo_sanguineo"`
	SlugPublico      uuid.UUID         `json:"slug_publico"`
	SenhaPublicaHash string            `json:"-"`
	Ativo            bool              `json:"ativo"`
	BloqueadoAte     *time.Time        `json:"bloqueado_ate,omitempty"`
	CriadoEm         time.Time         `json:"criado_em"`
	AtualizadoEm     time.Time         `json:"atualizado_em"`
}

// ---------- Clinical data -----------------------------------------------------

type ContatoEmergencia struct {
	ID         uuid.UUID `json:"id"`
	PerfilID   uuid.UUID `json:"perfil_id"`
	Nome       string    `json:"nome"`
	Parentesco string    `json:"parentesco,omitempty"`
	Telefone   string    `json:"telefone"`
	Principal  bool      `json:"principal"`
}

type Alergia struct {
	ID        uuid.UUID `json:"id"`
	PerfilID  uuid.UUID `json:"perfil_id"`
	Descricao string    `json:"descricao"`
	Gravidade string    `json:"gravidade,omitempty"`
}

type Medicamento struct {
	ID         uuid.UUID `json:"id"`
	PerfilID   uuid.UUID `json:"perfil_id"`
	Nome       string    `json:"nome"`
	Dosagem    string    `json:"dosagem,omitempty"`
	Frequencia string    `json:"frequencia,omitempty"`
}

type Doenca struct {
	ID         uuid.UUID `json:"id"`
	PerfilID   uuid.UUID `json:"perfil_id"`
	Descricao  string    `json:"descricao"`
	Observacao string    `json:"observacao,omitempty"`
}

type Cirurgia struct {
	ID            uuid.UUID `json:"id"`
	PerfilID      uuid.UUID `json:"perfil_id"`
	Descricao     string    `json:"descricao"`
	AnoRealizacao *int16    `json:"ano_realizacao,omitempty"`
}

// ---------- Complete profile (API response) -----------------------------------

type PerfilCompleto struct {
	Perfil
	Contatos     []ContatoEmergencia `json:"contatos_emergencia"`
	Alergias     []Alergia           `json:"alergias"`
	Medicamentos []Medicamento       `json:"medicamentos"`
	Doencas      []Doenca            `json:"doencas"`
	Cirurgias    []Cirurgia          `json:"cirurgias"`
}

// ---------- Requests ----------------------------------------------------------

type RegistroRequest struct {
	Email         string `json:"email"`
	Senha         string `json:"senha"`
	Consentimento bool   `json:"consentimento"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type PerfilRequest struct {
	Nome          string               `json:"nome"`
	Sobrenome     string               `json:"sobrenome"`
	Sexo          SexoTipo             `json:"sexo"`
	TipoSanguineo TipoSanguineoTipo    `json:"tipo_sanguineo"`
	SenhaPublica  string               `json:"senha_publica"`
	Contatos      []ContatoRequest     `json:"contatos_emergencia"`
	Alergias      []AlergiaRequest     `json:"alergias"`
	Medicamentos  []MedicamentoRequest `json:"medicamentos"`
	Doencas       []DoencaRequest      `json:"doencas"`
	Cirurgias     []CirurgiaRequest    `json:"cirurgias"`
}

type ContatoRequest struct {
	Nome       string `json:"nome"`
	Parentesco string `json:"parentesco"`
	Telefone   string `json:"telefone"`
	Principal  bool   `json:"principal"`
}

type AlergiaRequest struct {
	Descricao string `json:"descricao"`
	Gravidade string `json:"gravidade"`
}

type MedicamentoRequest struct {
	Nome       string `json:"nome"`
	Dosagem    string `json:"dosagem"`
	Frequencia string `json:"frequencia"`
}

type DoencaRequest struct {
	Descricao  string `json:"descricao"`
	Observacao string `json:"observacao"`
}

type CirurgiaRequest struct {
	Descricao     string `json:"descricao"`
	AnoRealizacao *int16 `json:"ano_realizacao"`
}

type AcessoPublicoRequest struct {
	Senha string `json:"senha"`
}

// ---------- Responses ---------------------------------------------------------

type TokenResponse struct {
	Token string `json:"token"`
}

type ErroResponse struct {
	Erro string `json:"erro"`
}
