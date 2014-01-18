package entropy

type IValidator interface {
	Verify(value string) (bool, string)
}
