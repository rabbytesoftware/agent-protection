package dispatch

type Command interface {
	Run(
		target string,
	) (code int)
}
