package lang

type Compiler struct {
	RawFiles map[string]string
}

func (me *Compiler) Compile() error {
	return nil
}

func (me *Compiler) Generate() error {
	return nil
}
