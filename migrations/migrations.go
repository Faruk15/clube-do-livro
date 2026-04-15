// Package migrations embute os arquivos SQL numerados aplicados
// automaticamente pelo servidor na inicialização.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
