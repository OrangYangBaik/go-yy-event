package dtos

type Member struct {
	Email      string `json:"email" validate:"required"`
	Nama       string `json:"nama" validate:"required"`
	NoTelp     string `json:"no_telp" validate:"required"`
	Company    string `json:"company" validate:"required"`
	Jabatan    string `json:"jabatan" validate:"required"`
	Is_present bool   `json:"is_present"`
}

type Presensi struct {
	Email string `json:"email" validate:"required"`
}
