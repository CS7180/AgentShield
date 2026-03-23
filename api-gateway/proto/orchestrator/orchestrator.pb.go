// Manually maintained proto stubs.
// IMPORTANT: Run `make proto` to regenerate with protoc before enabling the gRPC orchestrator.
// These stubs compile correctly but will panic on actual gRPC serialization.
// Safe to use with ORCHESTRATOR_ENABLED=false (the default).
//
// source: proto/orchestrator/orchestrator.proto

package orchestrator

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StartScanRequest struct {
	ScanId         string   `protobuf:"bytes,1,opt,name=scan_id,json=scanId,proto3" json:"scan_id,omitempty"`
	TargetEndpoint string   `protobuf:"bytes,2,opt,name=target_endpoint,json=targetEndpoint,proto3" json:"target_endpoint,omitempty"`
	Mode           string   `protobuf:"bytes,3,opt,name=mode,proto3" json:"mode,omitempty"`
	AttackTypes    []string `protobuf:"bytes,4,rep,name=attack_types,json=attackTypes,proto3" json:"attack_types,omitempty"`
}

func (*StartScanRequest) ProtoMessage()                         {}
func (*StartScanRequest) ProtoReflect() protoreflect.Message   { return nil } //nolint:nilnil
func (x *StartScanRequest) Reset()                             { *x = StartScanRequest{} }
func (x *StartScanRequest) String() string                     { return x.ScanId }

type StartScanResponse struct {
	Accepted bool   `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
	Message  string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (*StartScanResponse) ProtoMessage()                         {}
func (*StartScanResponse) ProtoReflect() protoreflect.Message   { return nil } //nolint:nilnil
func (x *StartScanResponse) Reset()                             { *x = StartScanResponse{} }
func (x *StartScanResponse) String() string                     { return "" }

type StopScanRequest struct {
	ScanId string `protobuf:"bytes,1,opt,name=scan_id,json=scanId,proto3" json:"scan_id,omitempty"`
}

func (*StopScanRequest) ProtoMessage()                         {}
func (*StopScanRequest) ProtoReflect() protoreflect.Message   { return nil } //nolint:nilnil
func (x *StopScanRequest) Reset()                             { *x = StopScanRequest{} }
func (x *StopScanRequest) String() string                     { return x.ScanId }

type StopScanResponse struct {
	Stopped bool   `protobuf:"varint,1,opt,name=stopped,proto3" json:"stopped,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (*StopScanResponse) ProtoMessage()                         {}
func (*StopScanResponse) ProtoReflect() protoreflect.Message   { return nil } //nolint:nilnil
func (x *StopScanResponse) Reset()                             { *x = StopScanResponse{} }
func (x *StopScanResponse) String() string                     { return "" }

type ScanStatusRequest struct {
	ScanId string `protobuf:"bytes,1,opt,name=scan_id,json=scanId,proto3" json:"scan_id,omitempty"`
}

func (*ScanStatusRequest) ProtoMessage()                         {}
func (*ScanStatusRequest) ProtoReflect() protoreflect.Message   { return nil } //nolint:nilnil
func (x *ScanStatusRequest) Reset()                             { *x = ScanStatusRequest{} }
func (x *ScanStatusRequest) String() string                     { return x.ScanId }

type ScanStatusResponse struct {
	ScanId   string `protobuf:"bytes,1,opt,name=scan_id,json=scanId,proto3" json:"scan_id,omitempty"`
	Status   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Progress int32  `protobuf:"varint,3,opt,name=progress,proto3" json:"progress,omitempty"`
}

func (*ScanStatusResponse) ProtoMessage()                         {}
func (*ScanStatusResponse) ProtoReflect() protoreflect.Message   { return nil } //nolint:nilnil
func (x *ScanStatusResponse) Reset()                             { *x = ScanStatusResponse{} }
func (x *ScanStatusResponse) String() string                     { return "" }
