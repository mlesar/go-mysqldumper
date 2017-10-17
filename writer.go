package dumper

type DumpWriter interface {
	Write(data string) error
}
