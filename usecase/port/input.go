package port

type InputPort interface {
	Package(*PackageParams) (*PackageResponse, error)
	Service(*ServiceParams) (*ServiceResponse, error)

	Describe(*DescribeParams) (*DescribeResponse, error)
	Show(*ShowParams) (*ShowResponse, error)

	Header(*HeaderParams) (*HeaderResponse, error)

	Call(*CallParams) (*CallResponse, error)
}

type CallParams struct {
	PkgName string
	SvcName string
	RPCName string
}

type CallResponse struct{}

type DescribeParams struct {
	PkgName string
	MsgName string
}

type DescribeResponse struct{}

type PackageParams struct {
	PkgName string
}

type PackageResponse struct{}

type ServiceParams struct {
	SvcName string
}

type ServiceResponse struct{}

type ShowParams struct {
	Type     string      // TODO: enum
	Showable interface{} // TODO: interface
}

type ShowResponse struct{}

type HeaderParams struct {
	Key string
	Val string
}

type HeaderResponse struct{}
