package port

type OutputPort interface {
	Package() (*PackageResponse, error)
	Service() (*ServiceResponse, error)
	Describe() (*DescribeResponse, error)
	Show() (*ShowResponse, error)
	Header() (*HeaderResponse, error)
	Call() (*CallResponse, error)
}
