package mysqldumper

type DumpWriter interface {
	Write(data string) error
}
