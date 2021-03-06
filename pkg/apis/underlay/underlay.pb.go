// Code generated by protoc-gen-go. DO NOT EDIT.
// source: underlay.proto

package underlay

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Target struct {
	Cluster              string   `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Namespace            string   `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	ApiVersion           string   `protobuf:"bytes,3,opt,name=apiVersion,proto3" json:"apiVersion,omitempty"`
	Kind                 string   `protobuf:"bytes,4,opt,name=kind,proto3" json:"kind,omitempty"`
	Name                 string   `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Target) Reset()         { *m = Target{} }
func (m *Target) String() string { return proto.CompactTextString(m) }
func (*Target) ProtoMessage()    {}
func (*Target) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{0}
}

func (m *Target) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Target.Unmarshal(m, b)
}
func (m *Target) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Target.Marshal(b, m, deterministic)
}
func (m *Target) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Target.Merge(m, src)
}
func (m *Target) XXX_Size() int {
	return xxx_messageInfo_Target.Size(m)
}
func (m *Target) XXX_DiscardUnknown() {
	xxx_messageInfo_Target.DiscardUnknown(m)
}

var xxx_messageInfo_Target proto.InternalMessageInfo

func (m *Target) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *Target) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *Target) GetApiVersion() string {
	if m != nil {
		return m.ApiVersion
	}
	return ""
}

func (m *Target) GetKind() string {
	if m != nil {
		return m.Kind
	}
	return ""
}

func (m *Target) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type NodeRef struct {
	Cluster              string   `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeRef) Reset()         { *m = NodeRef{} }
func (m *NodeRef) String() string { return proto.CompactTextString(m) }
func (*NodeRef) ProtoMessage()    {}
func (*NodeRef) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{1}
}

func (m *NodeRef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeRef.Unmarshal(m, b)
}
func (m *NodeRef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeRef.Marshal(b, m, deterministic)
}
func (m *NodeRef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeRef.Merge(m, src)
}
func (m *NodeRef) XXX_Size() int {
	return xxx_messageInfo_NodeRef.Size(m)
}
func (m *NodeRef) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeRef.DiscardUnknown(m)
}

var xxx_messageInfo_NodeRef proto.InternalMessageInfo

func (m *NodeRef) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *NodeRef) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type PolicyRule struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PolicyRule) Reset()         { *m = PolicyRule{} }
func (m *PolicyRule) String() string { return proto.CompactTextString(m) }
func (*PolicyRule) ProtoMessage()    {}
func (*PolicyRule) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{2}
}

func (m *PolicyRule) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PolicyRule.Unmarshal(m, b)
}
func (m *PolicyRule) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PolicyRule.Marshal(b, m, deterministic)
}
func (m *PolicyRule) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PolicyRule.Merge(m, src)
}
func (m *PolicyRule) XXX_Size() int {
	return xxx_messageInfo_PolicyRule.Size(m)
}
func (m *PolicyRule) XXX_DiscardUnknown() {
	xxx_messageInfo_PolicyRule.DiscardUnknown(m)
}

var xxx_messageInfo_PolicyRule proto.InternalMessageInfo

type DiscoverRequest struct {
	Rules                []*PolicyRule `protobuf:"bytes,1,rep,name=rules,proto3" json:"rules,omitempty"`
	PeerNodes            []string      `protobuf:"bytes,2,rep,name=peerNodes,proto3" json:"peerNodes,omitempty"`
	EligibleNodes        []string      `protobuf:"bytes,3,rep,name=eligibleNodes,proto3" json:"eligibleNodes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *DiscoverRequest) Reset()         { *m = DiscoverRequest{} }
func (m *DiscoverRequest) String() string { return proto.CompactTextString(m) }
func (*DiscoverRequest) ProtoMessage()    {}
func (*DiscoverRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{3}
}

func (m *DiscoverRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DiscoverRequest.Unmarshal(m, b)
}
func (m *DiscoverRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DiscoverRequest.Marshal(b, m, deterministic)
}
func (m *DiscoverRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DiscoverRequest.Merge(m, src)
}
func (m *DiscoverRequest) XXX_Size() int {
	return xxx_messageInfo_DiscoverRequest.Size(m)
}
func (m *DiscoverRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DiscoverRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DiscoverRequest proto.InternalMessageInfo

func (m *DiscoverRequest) GetRules() []*PolicyRule {
	if m != nil {
		return m.Rules
	}
	return nil
}

func (m *DiscoverRequest) GetPeerNodes() []string {
	if m != nil {
		return m.PeerNodes
	}
	return nil
}

func (m *DiscoverRequest) GetEligibleNodes() []string {
	if m != nil {
		return m.EligibleNodes
	}
	return nil
}

type DiscoverResponse struct {
	Offers               []*Offer `protobuf:"bytes,1,rep,name=offers,proto3" json:"offers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DiscoverResponse) Reset()         { *m = DiscoverResponse{} }
func (m *DiscoverResponse) String() string { return proto.CompactTextString(m) }
func (*DiscoverResponse) ProtoMessage()    {}
func (*DiscoverResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{4}
}

func (m *DiscoverResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DiscoverResponse.Unmarshal(m, b)
}
func (m *DiscoverResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DiscoverResponse.Marshal(b, m, deterministic)
}
func (m *DiscoverResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DiscoverResponse.Merge(m, src)
}
func (m *DiscoverResponse) XXX_Size() int {
	return xxx_messageInfo_DiscoverResponse.Size(m)
}
func (m *DiscoverResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DiscoverResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DiscoverResponse proto.InternalMessageInfo

func (m *DiscoverResponse) GetOffers() []*Offer {
	if m != nil {
		return m.Offers
	}
	return nil
}

type Offer struct {
	Id                   string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Node                 *NodeRef               `protobuf:"bytes,2,opt,name=node,proto3" json:"node,omitempty"`
	Cost                 int64                  `protobuf:"varint,3,opt,name=cost,proto3" json:"cost,omitempty"`
	Expires              *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=expires,proto3" json:"expires,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *Offer) Reset()         { *m = Offer{} }
func (m *Offer) String() string { return proto.CompactTextString(m) }
func (*Offer) ProtoMessage()    {}
func (*Offer) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{5}
}

func (m *Offer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Offer.Unmarshal(m, b)
}
func (m *Offer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Offer.Marshal(b, m, deterministic)
}
func (m *Offer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Offer.Merge(m, src)
}
func (m *Offer) XXX_Size() int {
	return xxx_messageInfo_Offer.Size(m)
}
func (m *Offer) XXX_DiscardUnknown() {
	xxx_messageInfo_Offer.DiscardUnknown(m)
}

var xxx_messageInfo_Offer proto.InternalMessageInfo

func (m *Offer) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Offer) GetNode() *NodeRef {
	if m != nil {
		return m.Node
	}
	return nil
}

func (m *Offer) GetCost() int64 {
	if m != nil {
		return m.Cost
	}
	return 0
}

func (m *Offer) GetExpires() *timestamppb.Timestamp {
	if m != nil {
		return m.Expires
	}
	return nil
}

type AllocateRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AllocateRequest) Reset()         { *m = AllocateRequest{} }
func (m *AllocateRequest) String() string { return proto.CompactTextString(m) }
func (*AllocateRequest) ProtoMessage()    {}
func (*AllocateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{6}
}

func (m *AllocateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AllocateRequest.Unmarshal(m, b)
}
func (m *AllocateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AllocateRequest.Marshal(b, m, deterministic)
}
func (m *AllocateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AllocateRequest.Merge(m, src)
}
func (m *AllocateRequest) XXX_Size() int {
	return xxx_messageInfo_AllocateRequest.Size(m)
}
func (m *AllocateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AllocateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AllocateRequest proto.InternalMessageInfo

func (m *AllocateRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type AllocateResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AllocateResponse) Reset()         { *m = AllocateResponse{} }
func (m *AllocateResponse) String() string { return proto.CompactTextString(m) }
func (*AllocateResponse) ProtoMessage()    {}
func (*AllocateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{7}
}

func (m *AllocateResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AllocateResponse.Unmarshal(m, b)
}
func (m *AllocateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AllocateResponse.Marshal(b, m, deterministic)
}
func (m *AllocateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AllocateResponse.Merge(m, src)
}
func (m *AllocateResponse) XXX_Size() int {
	return xxx_messageInfo_AllocateResponse.Size(m)
}
func (m *AllocateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AllocateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AllocateResponse proto.InternalMessageInfo

type ReleaseRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReleaseRequest) Reset()         { *m = ReleaseRequest{} }
func (m *ReleaseRequest) String() string { return proto.CompactTextString(m) }
func (*ReleaseRequest) ProtoMessage()    {}
func (*ReleaseRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{8}
}

func (m *ReleaseRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReleaseRequest.Unmarshal(m, b)
}
func (m *ReleaseRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReleaseRequest.Marshal(b, m, deterministic)
}
func (m *ReleaseRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReleaseRequest.Merge(m, src)
}
func (m *ReleaseRequest) XXX_Size() int {
	return xxx_messageInfo_ReleaseRequest.Size(m)
}
func (m *ReleaseRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReleaseRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReleaseRequest proto.InternalMessageInfo

func (m *ReleaseRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type MitigateRequest struct {
	Id                   []string      `protobuf:"bytes,1,rep,name=id,proto3" json:"id,omitempty"`
	Src                  *NodeRef      `protobuf:"bytes,2,opt,name=src,proto3" json:"src,omitempty"`
	Peer                 *NodeRef      `protobuf:"bytes,3,opt,name=peer,proto3" json:"peer,omitempty"`
	Rules                []*PolicyRule `protobuf:"bytes,4,rep,name=rules,proto3" json:"rules,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *MitigateRequest) Reset()         { *m = MitigateRequest{} }
func (m *MitigateRequest) String() string { return proto.CompactTextString(m) }
func (*MitigateRequest) ProtoMessage()    {}
func (*MitigateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{9}
}

func (m *MitigateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MitigateRequest.Unmarshal(m, b)
}
func (m *MitigateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MitigateRequest.Marshal(b, m, deterministic)
}
func (m *MitigateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MitigateRequest.Merge(m, src)
}
func (m *MitigateRequest) XXX_Size() int {
	return xxx_messageInfo_MitigateRequest.Size(m)
}
func (m *MitigateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MitigateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MitigateRequest proto.InternalMessageInfo

func (m *MitigateRequest) GetId() []string {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *MitigateRequest) GetSrc() *NodeRef {
	if m != nil {
		return m.Src
	}
	return nil
}

func (m *MitigateRequest) GetPeer() *NodeRef {
	if m != nil {
		return m.Peer
	}
	return nil
}

func (m *MitigateRequest) GetRules() []*PolicyRule {
	if m != nil {
		return m.Rules
	}
	return nil
}

type MitigateResponse struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MitigateResponse) Reset()         { *m = MitigateResponse{} }
func (m *MitigateResponse) String() string { return proto.CompactTextString(m) }
func (*MitigateResponse) ProtoMessage()    {}
func (*MitigateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b24729d943982f0, []int{10}
}

func (m *MitigateResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MitigateResponse.Unmarshal(m, b)
}
func (m *MitigateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MitigateResponse.Marshal(b, m, deterministic)
}
func (m *MitigateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MitigateResponse.Merge(m, src)
}
func (m *MitigateResponse) XXX_Size() int {
	return xxx_messageInfo_MitigateResponse.Size(m)
}
func (m *MitigateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MitigateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MitigateResponse proto.InternalMessageInfo

func (m *MitigateResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func init() {
	proto.RegisterType((*Target)(nil), "underlay.Target")
	proto.RegisterType((*NodeRef)(nil), "underlay.NodeRef")
	proto.RegisterType((*PolicyRule)(nil), "underlay.PolicyRule")
	proto.RegisterType((*DiscoverRequest)(nil), "underlay.DiscoverRequest")
	proto.RegisterType((*DiscoverResponse)(nil), "underlay.DiscoverResponse")
	proto.RegisterType((*Offer)(nil), "underlay.Offer")
	proto.RegisterType((*AllocateRequest)(nil), "underlay.AllocateRequest")
	proto.RegisterType((*AllocateResponse)(nil), "underlay.AllocateResponse")
	proto.RegisterType((*ReleaseRequest)(nil), "underlay.ReleaseRequest")
	proto.RegisterType((*MitigateRequest)(nil), "underlay.MitigateRequest")
	proto.RegisterType((*MitigateResponse)(nil), "underlay.MitigateResponse")
}

func init() { proto.RegisterFile("underlay.proto", fileDescriptor_1b24729d943982f0) }

var fileDescriptor_1b24729d943982f0 = []byte{
	// 559 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x54, 0x4b, 0x6f, 0xd3, 0x4c,
	0x14, 0x95, 0xed, 0xb4, 0x69, 0x6f, 0xbe, 0x2f, 0x49, 0x47, 0x08, 0x19, 0x83, 0x4a, 0x30, 0x20,
	0x2a, 0x16, 0x8e, 0x94, 0x22, 0xb1, 0xe8, 0x2a, 0x3c, 0x96, 0x3c, 0x64, 0x15, 0x16, 0xec, 0x1c,
	0xfb, 0xc6, 0x8c, 0x32, 0xf1, 0x98, 0x99, 0x31, 0xa2, 0x4b, 0x76, 0xfc, 0x05, 0xf8, 0xb5, 0x68,
	0xc6, 0x9e, 0xd8, 0x35, 0x29, 0xec, 0xc6, 0xe7, 0xbe, 0xce, 0x3d, 0xe7, 0xca, 0x30, 0xae, 0x8a,
	0x0c, 0x05, 0x4b, 0xae, 0xa2, 0x52, 0x70, 0xc5, 0xc9, 0x91, 0xfd, 0x0e, 0xee, 0xe6, 0x9c, 0xe7,
	0x0c, 0xe7, 0x06, 0x5f, 0x55, 0xeb, 0x39, 0x6e, 0x4b, 0xd5, 0xa4, 0x05, 0xf7, 0xfb, 0x41, 0x45,
	0xb7, 0x28, 0x55, 0xb2, 0x2d, 0xeb, 0x84, 0xf0, 0x87, 0x03, 0x87, 0x97, 0x89, 0xc8, 0x51, 0x11,
	0x1f, 0x86, 0x29, 0xab, 0xa4, 0x42, 0xe1, 0x3b, 0x33, 0xe7, 0xec, 0x38, 0xb6, 0x9f, 0xe4, 0x1e,
	0x1c, 0x17, 0xc9, 0x16, 0x65, 0x99, 0xa4, 0xe8, 0xbb, 0x26, 0xd6, 0x02, 0xe4, 0x14, 0x20, 0x29,
	0xe9, 0x47, 0x14, 0x92, 0xf2, 0xc2, 0xf7, 0x4c, 0xb8, 0x83, 0x10, 0x02, 0x83, 0x0d, 0x2d, 0x32,
	0x7f, 0x60, 0x22, 0xe6, 0xad, 0x31, 0xdd, 0xc0, 0x3f, 0xa8, 0x31, 0xfd, 0x0e, 0x9f, 0xc3, 0xf0,
	0x2d, 0xcf, 0x30, 0xc6, 0xf5, 0x5f, 0xa8, 0xd8, 0x42, 0xb7, 0x53, 0xf8, 0x1f, 0xc0, 0x7b, 0xce,
	0x68, 0x7a, 0x15, 0x57, 0x0c, 0xc3, 0xef, 0x0e, 0x4c, 0x5e, 0x51, 0x99, 0xf2, 0xaf, 0x28, 0x62,
	0xfc, 0x52, 0xa1, 0x54, 0xe4, 0x29, 0x1c, 0x88, 0x8a, 0xa1, 0xf4, 0x9d, 0x99, 0x77, 0x36, 0x5a,
	0xdc, 0x8a, 0x76, 0x6a, 0xb6, 0x85, 0x71, 0x9d, 0xa2, 0x97, 0x2d, 0x11, 0x85, 0xa6, 0x22, 0x7d,
	0x77, 0xe6, 0xe9, 0x65, 0x77, 0x00, 0x79, 0x04, 0xff, 0x23, 0xa3, 0x39, 0x5d, 0x31, 0xac, 0x33,
	0x3c, 0x93, 0x71, 0x1d, 0x0c, 0x2f, 0x60, 0xda, 0x52, 0x90, 0x25, 0x2f, 0x24, 0x92, 0x27, 0x70,
	0xc8, 0xd7, 0x6b, 0x14, 0x96, 0xc4, 0xa4, 0x25, 0xf1, 0x4e, 0xe3, 0x71, 0x13, 0xd6, 0x96, 0x1c,
	0x18, 0x84, 0x8c, 0xc1, 0xa5, 0x59, 0xa3, 0x80, 0x4b, 0x33, 0xf2, 0x18, 0x06, 0x05, 0xcf, 0xea,
	0xe5, 0x47, 0x8b, 0x93, 0xb6, 0x41, 0xa3, 0x5b, 0x6c, 0xc2, 0x5a, 0xa3, 0x94, 0x4b, 0x65, 0xac,
	0xf0, 0x62, 0xf3, 0x26, 0xcf, 0x60, 0x88, 0xdf, 0x4a, 0x2a, 0x50, 0x1a, 0x1f, 0x46, 0x8b, 0x20,
	0xaa, 0x4f, 0x23, 0xb2, 0xa7, 0x11, 0x5d, 0xda, 0xd3, 0x88, 0x6d, 0x6a, 0xf8, 0x00, 0x26, 0x4b,
	0xc6, 0x78, 0x9a, 0x28, 0xb4, 0x52, 0xf6, 0x38, 0x85, 0x04, 0xa6, 0x6d, 0x4a, 0xbd, 0x6a, 0x38,
	0x83, 0x71, 0x8c, 0x0c, 0x13, 0x79, 0x63, 0xd5, 0x4f, 0x07, 0x26, 0x6f, 0xa8, 0xa2, 0xf9, 0x9e,
	0xce, 0x5e, 0xb3, 0xed, 0x43, 0xf0, 0xa4, 0x48, 0x6f, 0x5e, 0x56, 0x47, 0xb5, 0x24, 0xda, 0x1c,
	0xb3, 0xeb, 0x7e, 0x49, 0x74, 0xb8, 0x3d, 0x80, 0xc1, 0x3f, 0x0f, 0x20, 0x0c, 0x61, 0xda, 0x52,
	0x6b, 0xcc, 0xeb, 0xf1, 0x5f, 0xfc, 0x72, 0x81, 0x7c, 0x68, 0x5a, 0xbc, 0xe4, 0x85, 0x12, 0x9c,
	0x31, 0x14, 0x64, 0x09, 0x47, 0xd6, 0x77, 0x72, 0xa7, 0x9d, 0xd1, 0x3b, 0xc7, 0x20, 0xd8, 0x17,
	0x6a, 0x26, 0x2d, 0xe1, 0xc8, 0xea, 0xd9, 0x6d, 0xd1, 0xb3, 0xa1, 0xdb, 0xa2, 0x2f, 0x3f, 0xb9,
	0x80, 0x61, 0x23, 0x3f, 0xf1, 0xdb, 0xb4, 0xeb, 0x8e, 0x04, 0xb7, 0xff, 0xf0, 0xff, 0xb5, 0xfe,
	0x6f, 0xe8, 0xf9, 0x76, 0xfb, 0xee, 0xfc, 0x9e, 0x59, 0xdd, 0xf9, 0x7d, 0xb1, 0x5e, 0x9c, 0xc3,
	0x69, 0x4e, 0xd5, 0xe7, 0x6a, 0x15, 0xa5, 0x7c, 0x3b, 0x57, 0x95, 0x28, 0x56, 0x55, 0xba, 0x61,
	0x38, 0xb7, 0x25, 0x9f, 0x4e, 0xca, 0x4d, 0x3e, 0x4f, 0x4a, 0x2a, 0x77, 0xd0, 0xea, 0xd0, 0xf0,
	0x38, 0xff, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x5c, 0x9c, 0xfc, 0xf0, 0xe9, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// UnderlayControllerClient is the client API for UnderlayController service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type UnderlayControllerClient interface {
	Discover(ctx context.Context, in *DiscoverRequest, opts ...grpc.CallOption) (*DiscoverResponse, error)
	Allocate(ctx context.Context, in *AllocateRequest, opts ...grpc.CallOption) (*AllocateResponse, error)
	Release(ctx context.Context, in *ReleaseRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Mitigate(ctx context.Context, in *MitigateRequest, opts ...grpc.CallOption) (*MitigateResponse, error)
}

type underlayControllerClient struct {
	cc *grpc.ClientConn
}

func NewUnderlayControllerClient(cc *grpc.ClientConn) UnderlayControllerClient {
	return &underlayControllerClient{cc}
}

func (c *underlayControllerClient) Discover(ctx context.Context, in *DiscoverRequest, opts ...grpc.CallOption) (*DiscoverResponse, error) {
	out := new(DiscoverResponse)
	err := c.cc.Invoke(ctx, "/underlay.UnderlayController/Discover", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *underlayControllerClient) Allocate(ctx context.Context, in *AllocateRequest, opts ...grpc.CallOption) (*AllocateResponse, error) {
	out := new(AllocateResponse)
	err := c.cc.Invoke(ctx, "/underlay.UnderlayController/Allocate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *underlayControllerClient) Release(ctx context.Context, in *ReleaseRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/underlay.UnderlayController/Release", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *underlayControllerClient) Mitigate(ctx context.Context, in *MitigateRequest, opts ...grpc.CallOption) (*MitigateResponse, error) {
	out := new(MitigateResponse)
	err := c.cc.Invoke(ctx, "/underlay.UnderlayController/Mitigate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UnderlayControllerServer is the server API for UnderlayController service.
type UnderlayControllerServer interface {
	Discover(context.Context, *DiscoverRequest) (*DiscoverResponse, error)
	Allocate(context.Context, *AllocateRequest) (*AllocateResponse, error)
	Release(context.Context, *ReleaseRequest) (*emptypb.Empty, error)
	Mitigate(context.Context, *MitigateRequest) (*MitigateResponse, error)
}

// UnimplementedUnderlayControllerServer can be embedded to have forward compatible implementations.
type UnimplementedUnderlayControllerServer struct {
}

func (*UnimplementedUnderlayControllerServer) Discover(ctx context.Context, req *DiscoverRequest) (*DiscoverResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Discover not implemented")
}
func (*UnimplementedUnderlayControllerServer) Allocate(ctx context.Context, req *AllocateRequest) (*AllocateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Allocate not implemented")
}
func (*UnimplementedUnderlayControllerServer) Release(ctx context.Context, req *ReleaseRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Release not implemented")
}
func (*UnimplementedUnderlayControllerServer) Mitigate(ctx context.Context, req *MitigateRequest) (*MitigateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Mitigate not implemented")
}

func RegisterUnderlayControllerServer(s *grpc.Server, srv UnderlayControllerServer) {
	s.RegisterService(&_UnderlayController_serviceDesc, srv)
}

func _UnderlayController_Discover_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiscoverRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UnderlayControllerServer).Discover(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/underlay.UnderlayController/Discover",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UnderlayControllerServer).Discover(ctx, req.(*DiscoverRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UnderlayController_Allocate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AllocateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UnderlayControllerServer).Allocate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/underlay.UnderlayController/Allocate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UnderlayControllerServer).Allocate(ctx, req.(*AllocateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UnderlayController_Release_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReleaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UnderlayControllerServer).Release(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/underlay.UnderlayController/Release",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UnderlayControllerServer).Release(ctx, req.(*ReleaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UnderlayController_Mitigate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MitigateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UnderlayControllerServer).Mitigate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/underlay.UnderlayController/Mitigate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UnderlayControllerServer).Mitigate(ctx, req.(*MitigateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _UnderlayController_serviceDesc = grpc.ServiceDesc{
	ServiceName: "underlay.UnderlayController",
	HandlerType: (*UnderlayControllerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Discover",
			Handler:    _UnderlayController_Discover_Handler,
		},
		{
			MethodName: "Allocate",
			Handler:    _UnderlayController_Allocate_Handler,
		},
		{
			MethodName: "Release",
			Handler:    _UnderlayController_Release_Handler,
		},
		{
			MethodName: "Mitigate",
			Handler:    _UnderlayController_Mitigate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "underlay.proto",
}
