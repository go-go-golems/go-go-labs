Title: message package - github.com/emersion/go-message - Go Packages

URL Source: https://pkg.go.dev/github.com/emersion/go-message

Markdown Content:
Skip to Main Content
Why Go
Why Go
Case Studies
Use Cases
Security Policy
Learn
Docs
Docs
Effective Go
Go User Manual
Standard library
Release Notes
Packages
Community
Community
Recorded Talks
Meetups 
Conferences 
Go blog
Go project
Get connected
Discover Packages
 
github.com/emersion/go-message
message
package
module
Version: v0.18.2 Latest 
Published: Sep 28, 2024 
License: MIT 
Imports: 11 
Imported by: 206
Details
 Valid go.mod file 
 Redistributable license 
 Tagged version 
 Stable version 
Learn more about best practices
Repository
github.com/emersion/go-message
Links
 Open Source Insights
README
Documentation
Source Files
Directories
Features
License
Overview
Index
Constants
Variables
Functions
Types
Examples
IsUnknownCharset(err)
IsUnknownEncoding(err)
type Entity
type Header
type HeaderFields
type MultipartReader
type ReadOptions
type UnknownCharsetError
type UnknownEncodingError
type WalkFunc
type Writer
New(header, body)
NewMultipart(header, parts)
Read(r)
ReadWithOptions(r, opts)
(e) MultipartReader()
(e) Walk(walkFunc)
(e) WriteTo(w)
HeaderFromMap(m)
(h) ContentDisposition()
(h) ContentType()
(h) Copy()
(h) Fields()
(h) FieldsByKey(k)
(h) SetContentDisposition(disp, params)
(h) SetContentType(t, params)
(h) SetText(k, v)
(h) Text(k)
(u) Error()
(u) Unwrap()
(u) Error()
(u) Unwrap()
CreateWriter(w, header)
(w) Close()
(w) CreatePart(header)
(w) Write(b)
 README ¶
go-message

A Go library for the Internet Message Format. It implements:

RFC 5322: Internet Message Format
RFC 2045, RFC 2046 and RFC 2047: Multipurpose Internet Mail Extensions
RFC 2183: Content-Disposition Header Field
Features
Streaming API
Automatic encoding and charset handling (to decode all charsets, add import _ "github.com/emersion/go-message/charset" to your application)
A mail subpackage to read and write mail messages
DKIM-friendly
A textproto subpackage that just implements the wire format
License

MIT

Expand ▾
 Documentation ¶
Overview ¶

Package message implements reading and writing multipurpose messages.

RFC 2045, RFC 2046 and RFC 2047 defines MIME, and RFC 2183 defines the Content-Disposition header field.

Add this import to your package if you want to handle most common charsets by default:

import (
	_ "github.com/emersion/go-message/charset"
)


Note, non-UTF-8 charsets are only supported when reading messages. Only UTF-8 is supported when writing messages.

Example (Transform) ¶
Index ¶
Variables
func IsUnknownCharset(err error) bool
func IsUnknownEncoding(err error) bool
type Entity
func New(header Header, body io.Reader) (*Entity, error)
func NewMultipart(header Header, parts []*Entity) (*Entity, error)
func Read(r io.Reader) (*Entity, error)
func ReadWithOptions(r io.Reader, opts *ReadOptions) (*Entity, error)
func (e *Entity) MultipartReader() MultipartReader
func (e *Entity) Walk(walkFunc WalkFunc) error
func (e *Entity) WriteTo(w io.Writer) error
type Header
func HeaderFromMap(m map[string][]string) Header
func (h *Header) ContentDisposition() (disp string, params map[string]string, err error)
func (h *Header) ContentType() (t string, params map[string]string, err error)
func (h *Header) Copy() Header
func (h *Header) Fields() HeaderFields
func (h *Header) FieldsByKey(k string) HeaderFields
func (h *Header) SetContentDisposition(disp string, params map[string]string)
func (h *Header) SetContentType(t string, params map[string]string)
func (h *Header) SetText(k, v string)
func (h *Header) Text(k string) (string, error)
type HeaderFields
type MultipartReader
type ReadOptions
type UnknownCharsetError
func (u UnknownCharsetError) Error() string
func (u UnknownCharsetError) Unwrap() error
type UnknownEncodingError
func (u UnknownEncodingError) Error() string
func (u UnknownEncodingError) Unwrap() error
type WalkFunc
type Writer
func CreateWriter(w io.Writer, header Header) (*Writer, error)
func (w *Writer) Close() error
func (w *Writer) CreatePart(header Header) (*Writer, error)
func (w *Writer) Write(b []byte) (int, error)
Examples ¶
Package (Transform)
Read
Writer
Constants ¶

This section is empty.

Variables ¶
View Source
var CharsetReader func(charset string, input io.Reader) (io.Reader, error)

CharsetReader, if non-nil, defines a function to generate charset-conversion readers, converting from the provided charset into UTF-8. Charsets are always lower-case. utf-8 and us-ascii charsets are handled by default. One of the the CharsetReader's result values must be non-nil.

Importing github.com/emersion/go-message/charset will set CharsetReader to a function that handles most common charsets. Alternatively, CharsetReader can be set to e.g. golang.org/x/net/html/charset.NewReaderLabel.

Functions ¶
func IsUnknownCharset ¶
added in v0.10.0
func IsUnknownCharset(err error) bool

IsUnknownCharset returns a boolean indicating whether the error is known to report that the charset advertised by the entity is unknown.

func IsUnknownEncoding ¶
func IsUnknownEncoding(err error) bool

IsUnknownEncoding returns a boolean indicating whether the error is known to report that the encoding advertised by the entity is unknown.

Types ¶
type Entity ¶
type Entity struct {
	Header Header    // The entity's header.
	Body   io.Reader // The decoded entity's body.
	// contains filtered or unexported fields
}

An Entity is either a whole message or a one of the parts in the body of a multipart entity.

func New ¶
func New(header Header, body io.Reader) (*Entity, error)

New makes a new message with the provided header and body. The entity's transfer encoding and charset are automatically decoded to UTF-8.

If the message uses an unknown transfer encoding or charset, New returns an error that verifies IsUnknownCharset, but also returns an Entity that can be read.

func NewMultipart ¶
func NewMultipart(header Header, parts []*Entity) (*Entity, error)

NewMultipart makes a new multipart message with the provided header and parts. The Content-Type header must begin with "multipart/".

If the message uses an unknown transfer encoding, NewMultipart returns an error that verifies IsUnknownCharset, but also returns an Entity that can be read.

func Read ¶
func Read(r io.Reader) (*Entity, error)

Read reads a message from r. The message's encoding and charset are automatically decoded to raw UTF-8. Note that this function only reads the message header.

If the message uses an unknown transfer encoding or charset, Read returns an error that verifies IsUnknownCharset or IsUnknownEncoding, but also returns an Entity that can be read.

Example ¶
func ReadWithOptions ¶
added in v0.16.0
func ReadWithOptions(r io.Reader, opts *ReadOptions) (*Entity, error)

ReadWithOptions see Read, but allows overriding some parameters with ReadOptions.

If the message uses an unknown transfer encoding or charset, ReadWithOptions returns an error that verifies IsUnknownCharset or IsUnknownEncoding, but also returns an Entity that can be read.

func (*Entity) MultipartReader ¶
func (e *Entity) MultipartReader() MultipartReader

MultipartReader returns a MultipartReader that reads parts from this entity's body. If this entity is not multipart, it returns nil.

func (*Entity) Walk ¶
added in v0.14.0
func (e *Entity) Walk(walkFunc WalkFunc) error

Walk walks the entity's multipart tree, calling walkFunc for each part in the tree, including the root entity.

Walk consumes the entity.

func (*Entity) WriteTo ¶
func (e *Entity) WriteTo(w io.Writer) error

WriteTo writes this entity's header and body to w.

type Header ¶
type Header struct {
	textproto.Header
}

A Header represents the key-value pairs in a message header.

func HeaderFromMap ¶
added in v0.15.0
func HeaderFromMap(m map[string][]string) Header

HeaderFromMap creates a header from a map of header fields.

This function is provided for interoperability with the standard library. If possible, ReadHeader should be used instead to avoid loosing information. The map representation looses the ordering of the fields, the capitalization of the header keys, and the whitespace of the original header.

func (*Header) ContentDisposition ¶
func (h *Header) ContentDisposition() (disp string, params map[string]string, err error)

ContentDisposition parses the Content-Disposition header field, as defined in RFC 2183.

func (*Header) ContentType ¶
func (h *Header) ContentType() (t string, params map[string]string, err error)

ContentType parses the Content-Type header field.

If no Content-Type is specified, it returns "text/plain".

func (*Header) Copy ¶
added in v0.14.0
func (h *Header) Copy() Header

Copy creates a stand-alone copy of the header.

func (*Header) Fields ¶
added in v0.10.4
func (h *Header) Fields() HeaderFields

Fields iterates over all the header fields.

The header may not be mutated while iterating, except using HeaderFields.Del.

func (*Header) FieldsByKey ¶
added in v0.10.4
func (h *Header) FieldsByKey(k string) HeaderFields

FieldsByKey iterates over all fields having the specified key.

The header may not be mutated while iterating, except using HeaderFields.Del.

func (*Header) SetContentDisposition ¶
func (h *Header) SetContentDisposition(disp string, params map[string]string)

SetContentDisposition formats the Content-Disposition header field, as defined in RFC 2183.

func (*Header) SetContentType ¶
func (h *Header) SetContentType(t string, params map[string]string)

SetContentType formats the Content-Type header field.

func (*Header) SetText ¶
added in v0.10.0
func (h *Header) SetText(k, v string)

SetText sets a plaintext header field.

func (*Header) Text ¶
added in v0.10.0
func (h *Header) Text(k string) (string, error)

Text parses a plaintext header field. The field charset is automatically decoded to UTF-8. If the header field's charset is unknown, the raw field value is returned and the error verifies IsUnknownCharset.

type HeaderFields ¶
added in v0.10.4
type HeaderFields interface {
	textproto.HeaderFields

	// Text parses the value of the current field as plaintext. The field
	// charset is decoded to UTF-8. If the header field's charset is unknown,
	// the raw field value is returned and the error verifies IsUnknownCharset.
	Text() (string, error)
}

HeaderFields iterates over header fields.

type MultipartReader ¶
type MultipartReader interface {
	io.Closer

	// NextPart returns the next part in the multipart or an error. When there are
	// no more parts, the error io.EOF is returned.
	//
	// Entity.Body must be read completely before the next call to NextPart,
	// otherwise it will be discarded.
	NextPart() (*Entity, error)
}

MultipartReader is an iterator over parts in a MIME multipart body.

type ReadOptions ¶
added in v0.16.0
type ReadOptions struct {
	// MaxHeaderBytes limits the maximum permissible size of a message header
	// block. If exceeded, an error will be returned.
	//
	// Set to -1 for no limit, set to 0 for the default value (1MB).
	MaxHeaderBytes int64
}

ReadOptions are options for ReadWithOptions.

type UnknownCharsetError ¶
added in v0.13.0
type UnknownCharsetError struct {
	// contains filtered or unexported fields
}
func (UnknownCharsetError) Error ¶
added in v0.13.0
func (u UnknownCharsetError) Error() string
func (UnknownCharsetError) Unwrap ¶
added in v0.13.0
func (u UnknownCharsetError) Unwrap() error
type UnknownEncodingError ¶
added in v0.13.0
type UnknownEncodingError struct {
	// contains filtered or unexported fields
}
func (UnknownEncodingError) Error ¶
added in v0.13.0
func (u UnknownEncodingError) Error() string
func (UnknownEncodingError) Unwrap ¶
added in v0.13.0
func (u UnknownEncodingError) Unwrap() error
type WalkFunc ¶
added in v0.14.0
type WalkFunc func(path []int, entity *Entity, err error) error

WalkFunc is the type of the function called for each part visited by Walk.

The path argument is a list of multipart indices leading to the part. The root part has a nil path.

If there was an encoding error walking to a part, the incoming error will describe the problem and the function can decide how to handle that error.

Unlike IMAP part paths, indices start from 0 (instead of 1) and a non-multipart message has a nil path (instead of {1}).

If an error is returned, processing stops.

type Writer ¶
type Writer struct {
	// contains filtered or unexported fields
}

Writer writes message entities.

If the message is not multipart, it should be used as a WriteCloser. Don't forget to call Close.

If the message is multipart, users can either use CreatePart to write child parts or Write to directly pipe a multipart message. In any case, Close must be called at the end.

Example ¶
func CreateWriter ¶
func CreateWriter(w io.Writer, header Header) (*Writer, error)

CreateWriter creates a new message writer to w. If header contains an encoding, data written to the Writer will automatically be encoded with it. The charset needs to be utf-8 or us-ascii.

func (*Writer) Close ¶
func (w *Writer) Close() error

Close implements io.Closer.

func (*Writer) CreatePart ¶
func (w *Writer) CreatePart(header Header) (*Writer, error)

CreatePart returns a Writer to a new part in this multipart entity. If this entity is not multipart, it fails. The body of the part should be written to the returned io.WriteCloser.

func (*Writer) Write ¶
func (w *Writer) Write(b []byte) (int, error)

Write implements io.Writer.

 Source Files ¶
View all Source files
charset.go
encoding.go
entity.go
header.go
message.go
multipart.go
writer.go
 Directories ¶
charset
	Package charset provides functions to decode and encode charsets.

mail
	Package mail implements reading and writing mail messages.

textproto
	Package textproto implements low-level manipulation of MIME messages.
Why Go
Use Cases
Case Studies
Get Started
Playground
Tour
Stack Overflow
Help
Packages
Standard Library
Sub-repositories
About Go Packages
About
Download
Blog
Issue Tracker
Release Notes
Brand Guidelines
Code of Conduct
Connect
Twitter
GitHub
Slack
r/golang
Meetup
Golang Weekly
Copyright
Terms of Service
Privacy Policy
Report an Issue

Theme Toggle

Shortcuts Modal

go.dev uses cookies from Google to deliver and enhance the quality of its services and to analyze traffic. Learn more.
Okay