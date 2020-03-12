// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/cloud/websecurityscanner/v1beta/finding_addon.proto

package websecurityscanner

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// ! Information about a vulnerability with an HTML.
type Form struct {
	// ! The URI where to send the form when it's submitted.
	ActionUri string `protobuf:"bytes,1,opt,name=action_uri,json=actionUri,proto3" json:"action_uri,omitempty"`
	// ! The names of form fields related to the vulnerability.
	Fields               []string `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Form) Reset()         { *m = Form{} }
func (m *Form) String() string { return proto.CompactTextString(m) }
func (*Form) ProtoMessage()    {}
func (*Form) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f8e8e70210d53, []int{0}
}

func (m *Form) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Form.Unmarshal(m, b)
}
func (m *Form) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Form.Marshal(b, m, deterministic)
}
func (m *Form) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Form.Merge(m, src)
}
func (m *Form) XXX_Size() int {
	return xxx_messageInfo_Form.Size(m)
}
func (m *Form) XXX_DiscardUnknown() {
	xxx_messageInfo_Form.DiscardUnknown(m)
}

var xxx_messageInfo_Form proto.InternalMessageInfo

func (m *Form) GetActionUri() string {
	if m != nil {
		return m.ActionUri
	}
	return ""
}

func (m *Form) GetFields() []string {
	if m != nil {
		return m.Fields
	}
	return nil
}

// Information reported for an outdated library.
type OutdatedLibrary struct {
	// The name of the outdated library.
	LibraryName string `protobuf:"bytes,1,opt,name=library_name,json=libraryName,proto3" json:"library_name,omitempty"`
	// The version number.
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	// URLs to learn more information about the vulnerabilities in the library.
	LearnMoreUrls        []string `protobuf:"bytes,3,rep,name=learn_more_urls,json=learnMoreUrls,proto3" json:"learn_more_urls,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OutdatedLibrary) Reset()         { *m = OutdatedLibrary{} }
func (m *OutdatedLibrary) String() string { return proto.CompactTextString(m) }
func (*OutdatedLibrary) ProtoMessage()    {}
func (*OutdatedLibrary) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f8e8e70210d53, []int{1}
}

func (m *OutdatedLibrary) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OutdatedLibrary.Unmarshal(m, b)
}
func (m *OutdatedLibrary) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OutdatedLibrary.Marshal(b, m, deterministic)
}
func (m *OutdatedLibrary) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OutdatedLibrary.Merge(m, src)
}
func (m *OutdatedLibrary) XXX_Size() int {
	return xxx_messageInfo_OutdatedLibrary.Size(m)
}
func (m *OutdatedLibrary) XXX_DiscardUnknown() {
	xxx_messageInfo_OutdatedLibrary.DiscardUnknown(m)
}

var xxx_messageInfo_OutdatedLibrary proto.InternalMessageInfo

func (m *OutdatedLibrary) GetLibraryName() string {
	if m != nil {
		return m.LibraryName
	}
	return ""
}

func (m *OutdatedLibrary) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *OutdatedLibrary) GetLearnMoreUrls() []string {
	if m != nil {
		return m.LearnMoreUrls
	}
	return nil
}

// Information regarding any resource causing the vulnerability such
// as JavaScript sources, image, audio files, etc.
type ViolatingResource struct {
	// The MIME type of this resource.
	ContentType string `protobuf:"bytes,1,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	// URL of this violating resource.
	ResourceUrl          string   `protobuf:"bytes,2,opt,name=resource_url,json=resourceUrl,proto3" json:"resource_url,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ViolatingResource) Reset()         { *m = ViolatingResource{} }
func (m *ViolatingResource) String() string { return proto.CompactTextString(m) }
func (*ViolatingResource) ProtoMessage()    {}
func (*ViolatingResource) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f8e8e70210d53, []int{2}
}

func (m *ViolatingResource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ViolatingResource.Unmarshal(m, b)
}
func (m *ViolatingResource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ViolatingResource.Marshal(b, m, deterministic)
}
func (m *ViolatingResource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ViolatingResource.Merge(m, src)
}
func (m *ViolatingResource) XXX_Size() int {
	return xxx_messageInfo_ViolatingResource.Size(m)
}
func (m *ViolatingResource) XXX_DiscardUnknown() {
	xxx_messageInfo_ViolatingResource.DiscardUnknown(m)
}

var xxx_messageInfo_ViolatingResource proto.InternalMessageInfo

func (m *ViolatingResource) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

func (m *ViolatingResource) GetResourceUrl() string {
	if m != nil {
		return m.ResourceUrl
	}
	return ""
}

// Information about vulnerable request parameters.
type VulnerableParameters struct {
	// The vulnerable parameter names.
	ParameterNames       []string `protobuf:"bytes,1,rep,name=parameter_names,json=parameterNames,proto3" json:"parameter_names,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VulnerableParameters) Reset()         { *m = VulnerableParameters{} }
func (m *VulnerableParameters) String() string { return proto.CompactTextString(m) }
func (*VulnerableParameters) ProtoMessage()    {}
func (*VulnerableParameters) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f8e8e70210d53, []int{3}
}

func (m *VulnerableParameters) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VulnerableParameters.Unmarshal(m, b)
}
func (m *VulnerableParameters) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VulnerableParameters.Marshal(b, m, deterministic)
}
func (m *VulnerableParameters) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VulnerableParameters.Merge(m, src)
}
func (m *VulnerableParameters) XXX_Size() int {
	return xxx_messageInfo_VulnerableParameters.Size(m)
}
func (m *VulnerableParameters) XXX_DiscardUnknown() {
	xxx_messageInfo_VulnerableParameters.DiscardUnknown(m)
}

var xxx_messageInfo_VulnerableParameters proto.InternalMessageInfo

func (m *VulnerableParameters) GetParameterNames() []string {
	if m != nil {
		return m.ParameterNames
	}
	return nil
}

// Information about vulnerable or missing HTTP Headers.
type VulnerableHeaders struct {
	// List of vulnerable headers.
	Headers []*VulnerableHeaders_Header `protobuf:"bytes,1,rep,name=headers,proto3" json:"headers,omitempty"`
	// List of missing headers.
	MissingHeaders       []*VulnerableHeaders_Header `protobuf:"bytes,2,rep,name=missing_headers,json=missingHeaders,proto3" json:"missing_headers,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *VulnerableHeaders) Reset()         { *m = VulnerableHeaders{} }
func (m *VulnerableHeaders) String() string { return proto.CompactTextString(m) }
func (*VulnerableHeaders) ProtoMessage()    {}
func (*VulnerableHeaders) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f8e8e70210d53, []int{4}
}

func (m *VulnerableHeaders) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VulnerableHeaders.Unmarshal(m, b)
}
func (m *VulnerableHeaders) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VulnerableHeaders.Marshal(b, m, deterministic)
}
func (m *VulnerableHeaders) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VulnerableHeaders.Merge(m, src)
}
func (m *VulnerableHeaders) XXX_Size() int {
	return xxx_messageInfo_VulnerableHeaders.Size(m)
}
func (m *VulnerableHeaders) XXX_DiscardUnknown() {
	xxx_messageInfo_VulnerableHeaders.DiscardUnknown(m)
}

var xxx_messageInfo_VulnerableHeaders proto.InternalMessageInfo

func (m *VulnerableHeaders) GetHeaders() []*VulnerableHeaders_Header {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *VulnerableHeaders) GetMissingHeaders() []*VulnerableHeaders_Header {
	if m != nil {
		return m.MissingHeaders
	}
	return nil
}

// Describes a HTTP Header.
type VulnerableHeaders_Header struct {
	// Header name.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Header value.
	Value                string   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VulnerableHeaders_Header) Reset()         { *m = VulnerableHeaders_Header{} }
func (m *VulnerableHeaders_Header) String() string { return proto.CompactTextString(m) }
func (*VulnerableHeaders_Header) ProtoMessage()    {}
func (*VulnerableHeaders_Header) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f8e8e70210d53, []int{4, 0}
}

func (m *VulnerableHeaders_Header) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VulnerableHeaders_Header.Unmarshal(m, b)
}
func (m *VulnerableHeaders_Header) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VulnerableHeaders_Header.Marshal(b, m, deterministic)
}
func (m *VulnerableHeaders_Header) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VulnerableHeaders_Header.Merge(m, src)
}
func (m *VulnerableHeaders_Header) XXX_Size() int {
	return xxx_messageInfo_VulnerableHeaders_Header.Size(m)
}
func (m *VulnerableHeaders_Header) XXX_DiscardUnknown() {
	xxx_messageInfo_VulnerableHeaders_Header.DiscardUnknown(m)
}

var xxx_messageInfo_VulnerableHeaders_Header proto.InternalMessageInfo

func (m *VulnerableHeaders_Header) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *VulnerableHeaders_Header) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

// Information reported for an XSS.
type Xss struct {
	// Stack traces leading to the point where the XSS occurred.
	StackTraces []string `protobuf:"bytes,1,rep,name=stack_traces,json=stackTraces,proto3" json:"stack_traces,omitempty"`
	// An error message generated by a javascript breakage.
	ErrorMessage         string   `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Xss) Reset()         { *m = Xss{} }
func (m *Xss) String() string { return proto.CompactTextString(m) }
func (*Xss) ProtoMessage()    {}
func (*Xss) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f8e8e70210d53, []int{5}
}

func (m *Xss) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Xss.Unmarshal(m, b)
}
func (m *Xss) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Xss.Marshal(b, m, deterministic)
}
func (m *Xss) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Xss.Merge(m, src)
}
func (m *Xss) XXX_Size() int {
	return xxx_messageInfo_Xss.Size(m)
}
func (m *Xss) XXX_DiscardUnknown() {
	xxx_messageInfo_Xss.DiscardUnknown(m)
}

var xxx_messageInfo_Xss proto.InternalMessageInfo

func (m *Xss) GetStackTraces() []string {
	if m != nil {
		return m.StackTraces
	}
	return nil
}

func (m *Xss) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}

func init() {
	proto.RegisterType((*Form)(nil), "google.cloud.websecurityscanner.v1beta.Form")
	proto.RegisterType((*OutdatedLibrary)(nil), "google.cloud.websecurityscanner.v1beta.OutdatedLibrary")
	proto.RegisterType((*ViolatingResource)(nil), "google.cloud.websecurityscanner.v1beta.ViolatingResource")
	proto.RegisterType((*VulnerableParameters)(nil), "google.cloud.websecurityscanner.v1beta.VulnerableParameters")
	proto.RegisterType((*VulnerableHeaders)(nil), "google.cloud.websecurityscanner.v1beta.VulnerableHeaders")
	proto.RegisterType((*VulnerableHeaders_Header)(nil), "google.cloud.websecurityscanner.v1beta.VulnerableHeaders.Header")
	proto.RegisterType((*Xss)(nil), "google.cloud.websecurityscanner.v1beta.Xss")
}

func init() {
	proto.RegisterFile("google/cloud/websecurityscanner/v1beta/finding_addon.proto", fileDescriptor_ed5f8e8e70210d53)
}

var fileDescriptor_ed5f8e8e70210d53 = []byte{
	// 526 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0x51, 0x6f, 0xd3, 0x3c,
	0x14, 0x55, 0xbb, 0x7d, 0x9d, 0xea, 0x6e, 0xab, 0x6a, 0x4d, 0x9f, 0xa2, 0x09, 0xa4, 0x11, 0xa4,
	0x32, 0xf1, 0x90, 0x88, 0xf1, 0x06, 0x42, 0xc0, 0x90, 0x06, 0x0f, 0x14, 0xaa, 0x6e, 0x2d, 0x63,
	0xaa, 0x14, 0x39, 0xc9, 0x5d, 0xb0, 0x70, 0xec, 0xe8, 0xda, 0x29, 0xea, 0x9f, 0xe0, 0x57, 0xf1,
	0xc4, 0xaf, 0x42, 0xb1, 0x9d, 0x82, 0xd4, 0x07, 0xf6, 0xc0, 0x53, 0xee, 0x3d, 0xf7, 0x9e, 0x73,
	0xe2, 0x6b, 0x5f, 0xf2, 0xac, 0x50, 0xaa, 0x10, 0x10, 0x67, 0x42, 0xd5, 0x79, 0xfc, 0x0d, 0x52,
	0x0d, 0x59, 0x8d, 0xdc, 0xac, 0x75, 0xc6, 0xa4, 0x04, 0x8c, 0x57, 0x4f, 0x52, 0x30, 0x2c, 0xbe,
	0xe5, 0x32, 0xe7, 0xb2, 0x48, 0x58, 0x9e, 0x2b, 0x19, 0x55, 0xa8, 0x8c, 0xa2, 0x63, 0xc7, 0x8d,
	0x2c, 0x37, 0xda, 0xe6, 0x46, 0x8e, 0x7b, 0x7c, 0xcf, 0x7b, 0xb0, 0x8a, 0xc7, 0x4c, 0x4a, 0x65,
	0x98, 0xe1, 0x4a, 0x6a, 0xa7, 0x12, 0xbe, 0x20, 0xbb, 0x17, 0x0a, 0x4b, 0x7a, 0x9f, 0x10, 0x96,
	0x35, 0x85, 0xa4, 0x46, 0x1e, 0x74, 0x4e, 0x3a, 0xa7, 0xfd, 0x59, 0xdf, 0x21, 0x73, 0xe4, 0xf4,
	0x7f, 0xd2, 0xbb, 0xe5, 0x20, 0x72, 0x1d, 0x74, 0x4f, 0x76, 0x4e, 0xfb, 0x33, 0x9f, 0x85, 0x2b,
	0x32, 0xfc, 0x58, 0x9b, 0x9c, 0x19, 0xc8, 0xdf, 0xf3, 0x14, 0x19, 0xae, 0xe9, 0x03, 0xb2, 0x2f,
	0x5c, 0x98, 0x48, 0x56, 0x82, 0xd7, 0x1a, 0x78, 0xec, 0x03, 0x2b, 0x81, 0x06, 0x64, 0x6f, 0x05,
	0xa8, 0xb9, 0x92, 0x41, 0xd7, 0x56, 0xdb, 0x94, 0x8e, 0xc9, 0x50, 0x00, 0x43, 0x99, 0x94, 0x0a,
	0x21, 0xa9, 0x51, 0xe8, 0x60, 0xc7, 0x1a, 0x1e, 0x58, 0x78, 0xa2, 0x10, 0xe6, 0x28, 0x74, 0xf8,
	0x99, 0x8c, 0x16, 0x5c, 0x09, 0x66, 0xb8, 0x2c, 0x66, 0xa0, 0x55, 0x8d, 0x19, 0x34, 0xce, 0x99,
	0x92, 0x06, 0xa4, 0x49, 0xcc, 0xba, 0xda, 0x38, 0x7b, 0xec, 0x6a, 0x5d, 0xd9, 0x16, 0xf4, 0xed,
	0x8d, 0xba, 0xb7, 0x1f, 0xb4, 0xd8, 0x1c, 0x45, 0xf8, 0x92, 0x1c, 0x2d, 0x6a, 0x21, 0x01, 0x59,
	0x2a, 0x60, 0xca, 0x90, 0x95, 0x60, 0x00, 0x35, 0x7d, 0x44, 0x86, 0x55, 0x9b, 0xd9, 0x93, 0xe9,
	0xa0, 0x63, 0x7f, 0xed, 0x70, 0x03, 0x37, 0x87, 0xd3, 0xe1, 0xf7, 0x2e, 0x19, 0xfd, 0x56, 0x78,
	0x07, 0x2c, 0x6f, 0xe8, 0x37, 0x64, 0xef, 0x8b, 0x0b, 0x2d, 0x6d, 0x70, 0xf6, 0x2a, 0xba, 0xdb,
	0x05, 0x46, 0x5b, 0x5a, 0x91, 0xfb, 0xce, 0x5a, 0x41, 0xca, 0xc9, 0xb0, 0xe4, 0x5a, 0x37, 0x2f,
	0xa4, 0xf5, 0xe8, 0xfe, 0x23, 0x8f, 0x43, 0x2f, 0xec, 0xe1, 0xe3, 0x33, 0xd2, 0x73, 0x21, 0xa5,
	0x64, 0xf7, 0x8f, 0xfb, 0xb5, 0x31, 0x3d, 0x22, 0xff, 0xad, 0x98, 0xa8, 0xc1, 0xcf, 0xd5, 0x25,
	0xe1, 0x84, 0xec, 0x5c, 0x6b, 0xdd, 0xcc, 0x5e, 0x1b, 0x96, 0x7d, 0x4d, 0x0c, 0xb2, 0x6c, 0x33,
	0xbd, 0x81, 0xc5, 0xae, 0x2c, 0x44, 0x1f, 0x92, 0x03, 0x40, 0x54, 0x98, 0x94, 0xa0, 0x35, 0x2b,
	0x5a, 0x9d, 0x7d, 0x0b, 0x4e, 0x1c, 0x76, 0xfe, 0xa3, 0x43, 0x1e, 0x67, 0xaa, 0xbc, 0xe3, 0xd1,
	0xce, 0x47, 0x17, 0x6e, 0x79, 0x5e, 0x37, 0xbb, 0x33, 0x6d, 0x1e, 0xfd, 0xb4, 0x73, 0x73, 0xed,
	0xc9, 0x85, 0x12, 0x4c, 0x16, 0x91, 0xc2, 0x22, 0x2e, 0x40, 0xda, 0x95, 0x88, 0x5d, 0x89, 0x55,
	0x5c, 0xff, 0x6d, 0x2f, 0x9f, 0x6f, 0x57, 0x7e, 0x76, 0xc7, 0x6f, 0x2d, 0x7f, 0xf9, 0xa6, 0xe1,
	0x2e, 0x3f, 0x41, 0x7a, 0xe9, 0x3b, 0x2e, 0x5d, 0xc7, 0x72, 0x61, 0xb9, 0x69, 0xcf, 0xba, 0x3d,
	0xfd, 0x15, 0x00, 0x00, 0xff, 0xff, 0x62, 0x6a, 0xcb, 0x5a, 0x04, 0x04, 0x00, 0x00,
}
