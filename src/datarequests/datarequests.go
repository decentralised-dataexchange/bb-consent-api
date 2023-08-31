package datarequests

// Data Request type and status const
const (
	DataRequestMaxComments = 14

	DataRequestTypeDelete   = 1
	DataRequestTypeDownload = 2
	DataRequestTypeUpdate   = 3

	DataRequestStatusInitiated              = 1
	DataRequestStatusAcknowledged           = 2
	DataRequestStatusProcessedWithoutAction = 6
	DataRequestStatusProcessedWithAction    = 7
	DataRequestStatusUserCancelled          = 8
)

type iDString struct {
	ID  int
	Str string
}

// Note: Dont change the ID(s) if new type is needed then add at the end

// StatusTypes Array of id and string
var StatusTypes = []iDString{
	iDString{ID: DataRequestStatusInitiated, Str: "Request initiated"},
	iDString{ID: DataRequestStatusAcknowledged, Str: "Request acknowledged"},
	iDString{ID: DataRequestStatusProcessedWithoutAction, Str: "Request processed without action"},
	iDString{ID: DataRequestStatusProcessedWithAction, Str: "Request processed with action"},
	iDString{ID: DataRequestStatusUserCancelled, Str: "Request cancelled by user"},
}
