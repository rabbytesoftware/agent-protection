package lint

const config = `version: "2"
run:
  relative-path-mode: wd
linters:
  default: none
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
    - ineffassign
    - staticcheck
    - unused
  exclusions:
    generated: disable
  settings:
    errcheck:
      check-blank: true
      check-type-assertions: true
    interfacebloat:
      max: 3
    staticcheck:
      checks:
        - SA4006
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
