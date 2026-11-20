package lint

const config = `version: "2"
linters:
  default: none
  exclusions:
    paths:
      - "docs/"
      - "generated/"
      - "gen/"
  enable:
    - gocyclo
    - gocognit
    - funlen
    - nestif
    - mnd
    - dupl
    - gochecknoglobals
    - gochecknoinits
    - unparam
    - errcheck
    - nakedret
    - interfacebloat
  settings:
    errcheck:
      check-blank: true
      check-type-assertions: true
    interfacebloat:
      max: 3
    nakedret:
      max-func-lines: 0
    gocyclo:
      min-complexity: 8
    gocognit:
      min-complexity: 12
    funlen:
      lines: 40
      statements: -1
    nestif:
      min-complexity: 2
    dupl:
      threshold: 150
`
