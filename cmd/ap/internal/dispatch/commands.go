package dispatch

type Commands struct {
	Check    func(root string) (code int)
	Fix      func(root string) (code int)
	Markdown func(path string) (code int)
	Guide    func() (code int)
}
