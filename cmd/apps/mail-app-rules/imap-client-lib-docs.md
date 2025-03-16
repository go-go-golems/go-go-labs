Title: imapclient package - github.com/emersion/go-imap/v2/imapclient - Go Packages

URL Source: https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient

Markdown Content:
Package imapclient implements an IMAP client.

#### Charset decoding [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#hdr-Charset_decoding "Go to Charset decoding")

By default, only basic charset decoding is performed. For non-UTF-8 decoding of message subjects and e-mail address names, users can set Options.WordDecoder. For instance, to use go-message's collection of charsets:

import (
	"mime"

	"github.com/emersion/go-message/charset"
)

options := &imapclient.Options{
	WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
}
client, err := imapclient.DialTLS("imap.example.org:993", options)

*   [type AppendCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand)
*   *   [func (cmd \*AppendCommand) Close() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand.Close)
    *   [func (cmd \*AppendCommand) Wait() (\*imap.AppendData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand.Wait)
    *   [func (cmd \*AppendCommand) Write(b \[\]byte) (int, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand.Write)
*   [type CapabilityCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CapabilityCommand)
*   *   [func (cmd \*CapabilityCommand) Wait() (imap.CapSet, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CapabilityCommand.Wait)
*   [type Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)
*   *   [func DialInsecure(address string, options \*Options) (\*Client, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#DialInsecure)
    *   [func DialStartTLS(address string, options \*Options) (\*Client, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#DialStartTLS)
    *   [func DialTLS(address string, options \*Options) (\*Client, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#DialTLS)
    *   [func New(conn net.Conn, options \*Options) \*Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#New)
    *   [func NewStartTLS(conn net.Conn, options \*Options) (\*Client, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#NewStartTLS)
*   *   [func (c \*Client) Append(mailbox string, size int64, options \*imap.AppendOptions) \*AppendCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Append)
    *   [func (c \*Client) Authenticate(saslClient sasl.Client) error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Authenticate)
    *   [func (c \*Client) Capability() \*CapabilityCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Capability)
    *   [func (c \*Client) Caps() imap.CapSet](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Caps)
    *   [func (c \*Client) Close() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Close)
    *   [func (c \*Client) Copy(numSet imap.NumSet, mailbox string) \*CopyCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Copy)
    *   [func (c \*Client) Create(mailbox string, options \*imap.CreateOptions) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Create)
    *   [func (c \*Client) Delete(mailbox string) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Delete)
    *   [func (c \*Client) Enable(caps ...imap.Cap) \*EnableCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Enable)
    *   [func (c \*Client) Expunge() \*ExpungeCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Expunge)
    *   [func (c \*Client) Fetch(numSet imap.NumSet, options \*imap.FetchOptions) \*FetchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Fetch)
    *   [func (c \*Client) GetACL(mailbox string) \*GetACLCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.GetACL)
    *   [func (c \*Client) GetMetadata(mailbox string, entries \[\]string, options \*GetMetadataOptions) \*GetMetadataCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.GetMetadata)
    *   [func (c \*Client) GetQuota(root string) \*GetQuotaCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.GetQuota)
    *   [func (c \*Client) GetQuotaRoot(mailbox string) \*GetQuotaRootCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.GetQuotaRoot)
    *   [func (c \*Client) ID(idData \*imap.IDData) \*IDCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.ID)
    *   [func (c \*Client) Idle() (\*IdleCommand, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Idle)
    *   [func (c \*Client) List(ref, pattern string, options \*imap.ListOptions) \*ListCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.List)
    *   [func (c \*Client) Login(username, password string) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Login)
    *   [func (c \*Client) Logout() \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Logout)
    *   [func (c \*Client) Mailbox() \*SelectedMailbox](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Mailbox)
    *   [func (c \*Client) Move(numSet imap.NumSet, mailbox string) \*MoveCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Move)
    *   [func (c \*Client) MyRights(mailbox string) \*MyRightsCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.MyRights)
    *   [func (c \*Client) Namespace() \*NamespaceCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Namespace)
    *   [func (c \*Client) Noop() \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Noop)
    *   [func (c \*Client) Rename(mailbox, newName string) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Rename)
    *   [func (c \*Client) Search(criteria \*imap.SearchCriteria, options \*imap.SearchOptions) \*SearchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Search)
    *   [func (c \*Client) Select(mailbox string, options \*imap.SelectOptions) \*SelectCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Select)
    *   [func (c \*Client) SetACL(mailbox string, ri imap.RightsIdentifier, rm imap.RightModification, ...) \*SetACLCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.SetACL)
    *   [func (c \*Client) SetMetadata(mailbox string, entries map\[string\]\*\[\]byte) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.SetMetadata)
    *   [func (c \*Client) SetQuota(root string, limits map\[imap.QuotaResourceType\]int64) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.SetQuota)
    *   [func (c \*Client) Sort(options \*SortOptions) \*SortCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Sort)
    *   [func (c \*Client) State() imap.ConnState](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.State)
    *   [func (c \*Client) Status(mailbox string, options \*imap.StatusOptions) \*StatusCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Status)
    *   [func (c \*Client) Store(numSet imap.NumSet, store \*imap.StoreFlags, options \*imap.StoreOptions) \*FetchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Store)
    *   [func (c \*Client) Subscribe(mailbox string) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Subscribe)
    *   [func (c \*Client) Thread(options \*ThreadOptions) \*ThreadCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Thread)
    *   [func (c \*Client) UIDExpunge(uids imap.UIDSet) \*ExpungeCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.UIDExpunge)
    *   [func (c \*Client) UIDSearch(criteria \*imap.SearchCriteria, options \*imap.SearchOptions) \*SearchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.UIDSearch)
    *   [func (c \*Client) UIDSort(options \*SortOptions) \*SortCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.UIDSort)
    *   [func (c \*Client) UIDThread(options \*ThreadOptions) \*ThreadCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.UIDThread)
    *   [func (c \*Client) Unauthenticate() \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Unauthenticate)
    *   [func (c \*Client) Unselect() \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Unselect)
    *   [func (c \*Client) UnselectAndExpunge() \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.UnselectAndExpunge)
    *   [func (c \*Client) Unsubscribe(mailbox string) \*Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.Unsubscribe)
    *   [func (c \*Client) WaitGreeting() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.WaitGreeting)
*   [type Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)
*   *   [func (cmd \*Command) Wait() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command.Wait)
*   [type CopyCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CopyCommand)
*   *   [func (cmd \*CopyCommand) Wait() (\*imap.CopyData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CopyCommand.Wait)
*   [type EnableCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#EnableCommand)
*   *   [func (cmd \*EnableCommand) Wait() (\*EnableData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#EnableCommand.Wait)
*   [type EnableData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#EnableData)
*   [type ExpungeCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand)
*   *   [func (cmd \*ExpungeCommand) Close() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand.Close)
    *   [func (cmd \*ExpungeCommand) Collect() (\[\]uint32, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand.Collect)
    *   [func (cmd \*ExpungeCommand) Next() uint32](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand.Next)
*   [type FetchBinarySectionBuffer](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchBinarySectionBuffer)
*   [type FetchBodySectionBuffer](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchBodySectionBuffer)
*   [type FetchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand)
*   *   [func (cmd \*FetchCommand) Close() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand.Close)
    *   [func (cmd \*FetchCommand) Collect() (\[\]\*FetchMessageBuffer, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand.Collect)
    *   [func (cmd \*FetchCommand) Next() \*FetchMessageData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand.Next)
*   [type FetchItemData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemData)
*   [type FetchItemDataBinarySection](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySection)
*   *   [func (dataItem \*FetchItemDataBinarySection) MatchCommand(item \*imap.FetchItemBinarySection) bool](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySection.MatchCommand)
*   [type FetchItemDataBinarySectionSize](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySectionSize)
*   *   [func (data \*FetchItemDataBinarySectionSize) MatchCommand(item \*imap.FetchItemBinarySectionSize) bool](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySectionSize.MatchCommand)
*   [type FetchItemDataBodySection](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBodySection)
*   *   [func (dataItem \*FetchItemDataBodySection) MatchCommand(item \*imap.FetchItemBodySection) bool](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBodySection.MatchCommand)
*   [type FetchItemDataBodyStructure](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBodyStructure)
*   [type FetchItemDataEnvelope](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataEnvelope)
*   [type FetchItemDataFlags](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataFlags)
*   [type FetchItemDataInternalDate](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataInternalDate)
*   [type FetchItemDataModSeq](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataModSeq)
*   [type FetchItemDataRFC822Size](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataRFC822Size)
*   [type FetchItemDataUID](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataUID)
*   [type FetchMessageBuffer](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageBuffer)
*   *   [func (buf \*FetchMessageBuffer) FindBinarySection(section \*imap.FetchItemBinarySection) \[\]byte](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageBuffer.FindBinarySection)
    *   [func (buf \*FetchMessageBuffer) FindBinarySectionSize(part \[\]int) (uint32, bool)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageBuffer.FindBinarySectionSize)
    *   [func (buf \*FetchMessageBuffer) FindBodySection(section \*imap.FetchItemBodySection) \[\]byte](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageBuffer.FindBodySection)
*   [type FetchMessageData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageData)
*   *   [func (data \*FetchMessageData) Collect() (\*FetchMessageBuffer, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageData.Collect)
    *   [func (data \*FetchMessageData) Next() FetchItemData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageData.Next)
*   [type GetACLCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetACLCommand)
*   *   [func (cmd \*GetACLCommand) Wait() (\*GetACLData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetACLCommand.Wait)
*   [type GetACLData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetACLData)
*   [type GetMetadataCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataCommand)
*   *   [func (cmd \*GetMetadataCommand) Wait() (\*GetMetadataData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataCommand.Wait)
*   [type GetMetadataData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataData)
*   [type GetMetadataDepth](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataDepth)
*   *   [func (depth GetMetadataDepth) String() string](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataDepth.String)
*   [type GetMetadataOptions](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataOptions)
*   [type GetQuotaCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaCommand)
*   *   [func (cmd \*GetQuotaCommand) Wait() (\*QuotaData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaCommand.Wait)
*   [type GetQuotaRootCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaRootCommand)
*   *   [func (cmd \*GetQuotaRootCommand) Wait() (\[\]QuotaData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaRootCommand.Wait)
*   [type IDCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IDCommand)
*   *   [func (r \*IDCommand) Wait() (\*imap.IDData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IDCommand.Wait)
*   [type IdleCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IdleCommand)
*   *   [func (cmd \*IdleCommand) Close() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IdleCommand.Close)
    *   [func (cmd \*IdleCommand) Wait() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IdleCommand.Wait)
*   [type ListCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand)
*   *   [func (cmd \*ListCommand) Close() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand.Close)
    *   [func (cmd \*ListCommand) Collect() (\[\]\*imap.ListData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand.Collect)
    *   [func (cmd \*ListCommand) Next() \*imap.ListData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand.Next)
*   [type MoveCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MoveCommand)
*   *   [func (cmd \*MoveCommand) Wait() (\*MoveData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MoveCommand.Wait)
*   [type MoveData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MoveData)
*   [type MyRightsCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MyRightsCommand)
*   *   [func (cmd \*MyRightsCommand) Wait() (\*MyRightsData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MyRightsCommand.Wait)
*   [type MyRightsData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MyRightsData)
*   [type NamespaceCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#NamespaceCommand)
*   *   [func (cmd \*NamespaceCommand) Wait() (\*imap.NamespaceData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#NamespaceCommand.Wait)
*   [type Options](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Options)
*   [type QuotaData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#QuotaData)
*   [type QuotaResourceData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#QuotaResourceData)
*   [type SearchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SearchCommand)
*   *   [func (cmd \*SearchCommand) Wait() (\*imap.SearchData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SearchCommand.Wait)
*   [type SelectCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectCommand)
*   *   [func (cmd \*SelectCommand) Wait() (\*imap.SelectData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectCommand.Wait)
*   [type SelectedMailbox](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectedMailbox)
*   [type SetACLCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SetACLCommand)
*   *   [func (cmd \*SetACLCommand) Wait() error](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SetACLCommand.Wait)
*   [type SortCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCommand)
*   *   [func (cmd \*SortCommand) Wait() (\[\]uint32, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCommand.Wait)
*   [type SortCriterion](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCriterion)
*   [type SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey)
*   [type SortOptions](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortOptions)
*   [type StatusCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#StatusCommand)
*   *   [func (cmd \*StatusCommand) Wait() (\*imap.StatusData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#StatusCommand.Wait)
*   [type ThreadCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadCommand)
*   *   [func (cmd \*ThreadCommand) Wait() (\[\]ThreadData, error)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadCommand.Wait)
*   [type ThreadData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadData)
*   [type ThreadOptions](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadOptions)
*   [type UnilateralDataHandler](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#UnilateralDataHandler)
*   [type UnilateralDataMailbox](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#UnilateralDataMailbox)

*   [Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client)
*   [Client (Pipelining)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client-Pipelining)
*   [Client.Append](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Append)
*   [Client.Authenticate (Oauth)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Authenticate-Oauth)
*   [Client.Fetch](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Fetch)
*   [Client.Fetch (ParseBody)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Fetch-ParseBody)
*   [Client.Fetch (StreamBody)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Fetch-StreamBody)
*   [Client.Idle](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Idle)
*   [Client.List (Stream)](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.List-Stream)
*   [Client.Search](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Search)
*   [Client.Status](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Status)
*   [Client.Store](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#example-Client.Store)

This section is empty.

This section is empty.

This section is empty.

#### type [AppendCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/append.go#L36) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand "Go to AppendCommand")

type AppendCommand struct {
	
}

AppendCommand is an APPEND command.

Callers must write the message contents, then call Close.

#### func (\*AppendCommand) [Close](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/append.go#L47) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand.Close "Go to AppendCommand.Close")

#### func (\*AppendCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/append.go#L56) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand.Wait "Go to AppendCommand.Wait")

func (cmd \*[AppendCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand)) Wait() (\*imap.AppendData, [error](https://pkg.go.dev/builtin#error))

#### func (\*AppendCommand) [Write](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/append.go#L43) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand.Write "Go to AppendCommand.Write")

#### type [CapabilityCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/capability.go#L30) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CapabilityCommand "Go to CapabilityCommand")

type CapabilityCommand struct {
	
}

CapabilityCommand is a CAPABILITY command.

#### func (\*CapabilityCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/capability.go#L35) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CapabilityCommand.Wait "Go to CapabilityCommand.Wait")

func (cmd \*[CapabilityCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CapabilityCommand)) Wait() (imap.CapSet, [error](https://pkg.go.dev/builtin#error))

Client is an IMAP client.

IMAP commands are exposed as methods. These methods will block until the command has been sent to the server, but won't block until the server sends a response. They return a command struct which can be used to wait for the server response. This can be used to execute multiple commands concurrently, however care must be taken to avoid ambiguities. See [RFC 9051 section 5.5](https://rfc-editor.org/rfc/rfc9051.html#section-5.5).

A client can be safely used from multiple goroutines, however this doesn't guarantee any command ordering and is subject to the same caveats as command pipelining (see above). Additionally, some commands (e.g. StartTLS, Authenticate, Idle) block the client during their execution.

DialInsecure connects to an IMAP server without any encryption at all.

DialStartTLS connects to an IMAP server with STARTTLS.

DialTLS connects to an IMAP server with implicit TLS.

New creates a new IMAP client.

This function doesn't perform I/O.

A nil options pointer is equivalent to a zero options value.

NewStartTLS creates a new IMAP client with STARTTLS.

A nil options pointer is equivalent to a zero options value.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Append(mailbox [string](https://pkg.go.dev/builtin#string), size [int64](https://pkg.go.dev/builtin#int64), options \*imap.AppendOptions) \*[AppendCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#AppendCommand)

Append sends an APPEND command.

The caller must call AppendCommand.Close.

The options are optional.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Authenticate(saslClient sasl.Client) [error](https://pkg.go.dev/builtin#error)

Authenticate sends an AUTHENTICATE command.

Unlike other commands, this method blocks until the SASL exchange completes.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Capability() \*[CapabilityCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CapabilityCommand)

Capability sends a CAPABILITY command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Caps() imap.CapSet

Caps returns the capabilities advertised by the server.

When the server hasn't sent the capability list, this method will request it and block until it's received. If the capabilities cannot be fetched, nil is returned.

Close immediately closes the connection.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Copy(numSet imap.NumSet, mailbox [string](https://pkg.go.dev/builtin#string)) \*[CopyCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CopyCommand)

Copy sends a COPY command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Create(mailbox [string](https://pkg.go.dev/builtin#string), options \*imap.CreateOptions) \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Create sends a CREATE command.

A nil options pointer is equivalent to a zero options value.

Delete sends a DELETE command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Enable(caps ...imap.Cap) \*[EnableCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#EnableCommand)

Enable sends an ENABLE command.

This command requires support for IMAP4rev2 or the ENABLE extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Expunge() \*[ExpungeCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand)

Expunge sends an EXPUNGE command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Fetch(numSet imap.NumSet, options \*imap.FetchOptions) \*[FetchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand)

Fetch sends a FETCH command.

The caller must fully consume the FetchCommand. A simple way to do so is to defer a call to FetchCommand.Close.

A nil options pointer is equivalent to a zero options value.

GetACL sends a GETACL command.

This command requires support for the ACL extension.

GetMetadata sends a GETMETADATA command.

This command requires support for the METADATA or METADATA-SERVER extension.

GetQuota sends a GETQUOTA command.

This command requires support for the QUOTA extension.

GetQuotaRoot sends a GETQUOTAROOT command.

This command requires support for the QUOTA extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) ID(idData \*imap.IDData) \*[IDCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IDCommand)

ID sends an ID command.

The ID command is introduced in [RFC 2971](https://rfc-editor.org/rfc/rfc2971.html). It requires support for the ID extension.

An example ID command:

ID ("name" "go-imap" "version" "1.0" "os" "Linux" "os-version" "7.9.4" "vendor" "Yahoo")

Idle sends an IDLE command.

Unlike other commands, this method blocks until the server acknowledges it. On success, the IDLE command is running and other commands cannot be sent. The caller must invoke IdleCommand.Close to stop IDLE and unblock the client.

This command requires support for IMAP4rev2 or the IDLE extension. The IDLE command is restarted automatically to avoid getting disconnected due to inactivity timeouts.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) List(ref, pattern [string](https://pkg.go.dev/builtin#string), options \*imap.ListOptions) \*[ListCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand)

List sends a LIST command.

The caller must fully consume the ListCommand. A simple way to do so is to defer a call to ListCommand.Close.

A nil options pointer is equivalent to a zero options value.

A non-zero options value requires support for IMAP4rev2 or the LIST-EXTENDED extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Login(username, password [string](https://pkg.go.dev/builtin#string)) \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Login sends a LOGIN command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Logout() \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Logout sends a LOGOUT command.

This command informs the server that the client is done with the connection.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Mailbox() \*[SelectedMailbox](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectedMailbox)

Mailbox returns the state of the currently selected mailbox.

If there is no currently selected mailbox, nil is returned.

The returned struct must not be mutated.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Move(numSet imap.NumSet, mailbox [string](https://pkg.go.dev/builtin#string)) \*[MoveCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MoveCommand)

Move sends a MOVE command.

If the server doesn't support IMAP4rev2 nor the MOVE extension, a fallback with COPY + STORE + EXPUNGE commands is used.

MyRights sends a MYRIGHTS command.

This command requires support for the ACL extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Namespace() \*[NamespaceCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#NamespaceCommand)

Namespace sends a NAMESPACE command.

This command requires support for IMAP4rev2 or the NAMESPACE extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Noop() \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Noop sends a NOOP command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Rename(mailbox, newName [string](https://pkg.go.dev/builtin#string)) \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Rename sends a RENAME command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Search(criteria \*imap.SearchCriteria, options \*imap.SearchOptions) \*[SearchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SearchCommand)

Search sends a SEARCH command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Select(mailbox [string](https://pkg.go.dev/builtin#string), options \*imap.SelectOptions) \*[SelectCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectCommand)

Select sends a SELECT or EXAMINE command.

A nil options pointer is equivalent to a zero options value.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) SetACL(mailbox [string](https://pkg.go.dev/builtin#string), ri imap.RightsIdentifier, rm imap.RightModification, rs imap.RightSet) \*[SetACLCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SetACLCommand)

SetACL sends a SETACL command.

This command requires support for the ACL extension.

SetMetadata sends a SETMETADATA command.

To remove an entry, set it to nil.

This command requires support for the METADATA or METADATA-SERVER extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) SetQuota(root [string](https://pkg.go.dev/builtin#string), limits map\[imap.QuotaResourceType\][int64](https://pkg.go.dev/builtin#int64)) \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

SetQuota sends a SETQUOTA command.

This command requires support for the SETQUOTA extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Sort(options \*[SortOptions](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortOptions)) \*[SortCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCommand)

Sort sends a SORT command.

This command requires support for the SORT extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) State() imap.ConnState

State returns the current connection state of the client.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Status(mailbox [string](https://pkg.go.dev/builtin#string), options \*imap.StatusOptions) \*[StatusCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#StatusCommand)

Status sends a STATUS command.

A nil options pointer is equivalent to a zero options value.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Store(numSet imap.NumSet, store \*imap.StoreFlags, options \*imap.StoreOptions) \*[FetchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand)

Store sends a STORE command.

Unless StoreFlags.Silent is set, the server will return the updated values.

A nil options pointer is equivalent to a zero options value.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Subscribe(mailbox [string](https://pkg.go.dev/builtin#string)) \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Subscribe sends a SUBSCRIBE command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Thread(options \*[ThreadOptions](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadOptions)) \*[ThreadCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadCommand)

Thread sends a THREAD command.

This command requires support for the THREAD extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) UIDExpunge(uids imap.UIDSet) \*[ExpungeCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand)

UIDExpunge sends a UID EXPUNGE command.

This command requires support for IMAP4rev2 or the UIDPLUS extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) UIDSearch(criteria \*imap.SearchCriteria, options \*imap.SearchOptions) \*[SearchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SearchCommand)

UIDSearch sends a UID SEARCH command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) UIDSort(options \*[SortOptions](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortOptions)) \*[SortCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCommand)

UIDSort sends a UID SORT command.

See Sort.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) UIDThread(options \*[ThreadOptions](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadOptions)) \*[ThreadCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadCommand)

UIDThread sends a UID THREAD command.

See Thread.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Unauthenticate() \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Unauthenticate sends an UNAUTHENTICATE command.

This command requires support for the UNAUTHENTICATE extension.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Unselect() \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Unselect sends an UNSELECT command.

This command requires support for IMAP4rev2 or the UNSELECT extension.

#### func (\*Client) [UnselectAndExpunge](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/select.go#L39) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client.UnselectAndExpunge "Go to Client.UnselectAndExpunge")

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) UnselectAndExpunge() \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

UnselectAndExpunge sends a CLOSE command.

CLOSE implicitly performs a silent EXPUNGE command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) Unsubscribe(mailbox [string](https://pkg.go.dev/builtin#string)) \*[Command](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command)

Subscribe sends an UNSUBSCRIBE command.

func (c \*[Client](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Client)) WaitGreeting() [error](https://pkg.go.dev/builtin#error)

WaitGreeting waits for the server's initial greeting.

#### type [Command](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/client.go#L1199) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command "Go to Command")

Command is a basic IMAP command.

#### func (\*Command) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/client.go#L1204) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#Command.Wait "Go to Command.Wait")

Wait blocks until the command has completed.

#### type [CopyCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/copy.go#L20) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CopyCommand "Go to CopyCommand")

type CopyCommand struct {
	
}

CopyCommand is a COPY command.

#### func (\*CopyCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/copy.go#L25) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CopyCommand.Wait "Go to CopyCommand.Wait")

func (cmd \*[CopyCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#CopyCommand)) Wait() (\*imap.CopyData, [error](https://pkg.go.dev/builtin#error))

#### type [EnableCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/enable.go#L56) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#EnableCommand "Go to EnableCommand")

type EnableCommand struct {
	
}

EnableCommand is an ENABLE command.

#### func (\*EnableCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/enable.go#L61) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#EnableCommand.Wait "Go to EnableCommand.Wait")

type EnableData struct {
	
	Caps imap.CapSet
}

EnableData is the data returned by the ENABLE command.

#### type [ExpungeCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/expunge.go#L47) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand "Go to ExpungeCommand")

type ExpungeCommand struct {
	
}

ExpungeCommand is an EXPUNGE command.

The caller must fully consume the ExpungeCommand. A simple way to do so is to defer a call to FetchCommand.Close.

#### func (\*ExpungeCommand) [Close](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/expunge.go#L64) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand.Close "Go to ExpungeCommand.Close")

Close releases the command.

Calling Close unblocks the IMAP client decoder and lets it read the next responses. Next will always return nil after Close.

#### func (\*ExpungeCommand) [Collect](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/expunge.go#L74) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand.Collect "Go to ExpungeCommand.Collect")

Collect accumulates expunged sequence numbers into a list.

This is equivalent to calling Next repeatedly and then Close.

#### func (\*ExpungeCommand) [Next](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/expunge.go#L56) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ExpungeCommand.Next "Go to ExpungeCommand.Next")

Next advances to the next expunged message sequence number.

On success, the message sequence number is returned. On error or if there are no more messages, 0 is returned. To check the error value, use Close.

type FetchBinarySectionBuffer struct {
	Section \*imap.FetchItemBinarySection
	Bytes   \[\][byte](https://pkg.go.dev/builtin#byte)
}

FetchBinarySectionBuffer is a buffer for the data returned by FetchItemBinarySection.

#### type [FetchBodySectionBuffer](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L484) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchBodySectionBuffer "Go to FetchBodySectionBuffer")

type FetchBodySectionBuffer struct {
	Section \*imap.FetchItemBodySection
	Bytes   \[\][byte](https://pkg.go.dev/builtin#byte)
}

FetchBodySectionBuffer is a buffer for the data returned by FetchItemBodySection.

#### type [FetchCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L150) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand "Go to FetchCommand")

type FetchCommand struct {
	
}

FetchCommand is a FETCH command.

#### func (\*FetchCommand) [Close](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L205) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand.Close "Go to FetchCommand.Close")

Close releases the command.

Calling Close unblocks the IMAP client decoder and lets it read the next responses. Next will always return nil after Close.

#### func (\*FetchCommand) [Collect](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L219) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand.Collect "Go to FetchCommand.Collect")

Collect accumulates message data into a list.

This method will read and store message contents in memory. This is acceptable when the message contents have a reasonable size, but may not be suitable when fetching e.g. attachments.

This is equivalent to calling Next repeatedly and then Close.

#### func (\*FetchCommand) [Next](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L193) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand.Next "Go to FetchCommand.Next")

func (cmd \*[FetchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchCommand)) Next() \*[FetchMessageData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageData)

Next advances to the next message.

On success, the message is returned. On error or if there are no more messages, nil is returned. To check the error value, use Close.

type FetchItemData interface {
	
}

FetchItemData contains a message's FETCH item data.

type FetchItemDataBinarySection struct {
	Section \*imap.FetchItemBinarySection
	Literal imap.LiteralReader
}

FetchItemDataBinarySection holds data returned by FETCH BINARY\[\].

Literal might be nil.

#### func (\*FetchItemDataBinarySection) [MatchCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L410) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySection.MatchCommand "Go to FetchItemDataBinarySection.MatchCommand")

func (dataItem \*[FetchItemDataBinarySection](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySection)) MatchCommand(item \*imap.FetchItemBinarySection) [bool](https://pkg.go.dev/builtin#bool)

MatchCommand checks whether a section returned by the server in a response is compatible with a section requested by the client in a command.

type FetchItemDataBinarySectionSize struct {
	Part \[\][int](https://pkg.go.dev/builtin#int)
	Size [uint32](https://pkg.go.dev/builtin#uint32)
}

FetchItemDataBinarySectionSize holds data returned by FETCH BINARY.SIZE\[\].

#### func (\*FetchItemDataBinarySectionSize) [MatchCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L469) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySectionSize.MatchCommand "Go to FetchItemDataBinarySectionSize.MatchCommand")

func (data \*[FetchItemDataBinarySectionSize](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySectionSize)) MatchCommand(item \*imap.FetchItemBinarySectionSize) [bool](https://pkg.go.dev/builtin#bool)

MatchCommand checks whether a section size returned by the server in a response is compatible with a section size requested by the client in a command.

#### type [FetchItemDataBodySection](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L373) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBodySection "Go to FetchItemDataBodySection")

type FetchItemDataBodySection struct {
	Section \*imap.FetchItemBodySection
	Literal imap.LiteralReader
}

FetchItemDataBodySection holds data returned by FETCH BODY\[\].

Literal might be nil.

#### func (\*FetchItemDataBodySection) [MatchCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L388) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBodySection.MatchCommand "Go to FetchItemDataBodySection.MatchCommand")

func (dataItem \*[FetchItemDataBodySection](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBodySection)) MatchCommand(item \*imap.FetchItemBodySection) [bool](https://pkg.go.dev/builtin#bool)

MatchCommand checks whether a section returned by the server in a response is compatible with a section requested by the client in a command.

#### type [FetchItemDataBodyStructure](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L451) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBodyStructure "Go to FetchItemDataBodyStructure")

type FetchItemDataBodyStructure struct {
	BodyStructure imap.BodyStructure
	IsExtended    [bool](https://pkg.go.dev/builtin#bool) 
}

FetchItemDataBodyStructure holds data returned by FETCH BODYSTRUCTURE or FETCH BODY.

type FetchItemDataEnvelope struct {
	Envelope \*imap.Envelope
}

FetchItemDataEnvelope holds data returned by FETCH ENVELOPE.

type FetchItemDataFlags struct {
	Flags \[\]imap.Flag
}

FetchItemDataFlags holds data returned by FETCH FLAGS.

type FetchItemDataInternalDate struct {
	Time [time](https://pkg.go.dev/time).[Time](https://pkg.go.dev/time#Time)
}

FetchItemDataInternalDate holds data returned by FETCH INTERNALDATE.

type FetchItemDataModSeq struct {
	ModSeq [uint64](https://pkg.go.dev/builtin#uint64)
}

FetchItemDataModSeq holds data returned by FETCH MODSEQ.

This requires the CONDSTORE extension.

type FetchItemDataRFC822Size struct {
	Size [int64](https://pkg.go.dev/builtin#int64)
}

FetchItemDataRFC822Size holds data returned by FETCH RFC822.SIZE.

type FetchItemDataUID struct {
	UID imap.UID
}

FetchItemDataUID holds data returned by FETCH UID.

type FetchMessageBuffer struct {
	SeqNum            [uint32](https://pkg.go.dev/builtin#uint32)
	Flags             \[\]imap.Flag
	Envelope          \*imap.Envelope
	InternalDate      [time](https://pkg.go.dev/time).[Time](https://pkg.go.dev/time#Time)
	RFC822Size        [int64](https://pkg.go.dev/builtin#int64)
	UID               imap.UID
	BodyStructure     imap.BodyStructure
	BodySection       \[\][FetchBodySectionBuffer](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchBodySectionBuffer)
	BinarySection     \[\][FetchBinarySectionBuffer](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchBinarySectionBuffer)
	BinarySectionSize \[\][FetchItemDataBinarySectionSize](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemDataBinarySectionSize)
	ModSeq            [uint64](https://pkg.go.dev/builtin#uint64) 
}

FetchMessageBuffer is a buffer for the data returned by FetchMessageData.

The SeqNum field is always populated. All remaining fields are optional.

func (buf \*[FetchMessageBuffer](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageBuffer)) FindBinarySection(section \*imap.FetchItemBinarySection) \[\][byte](https://pkg.go.dev/builtin#byte)

FindBinarySection returns the contents of a requested binary section.

If the binary section is not found, nil is returned.

FindBinarySectionSize returns a requested binary section size.

If the binary section size is not found, false is returned.

#### func (\*FetchMessageBuffer) [FindBodySection](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/fetch.go#L566) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageBuffer.FindBodySection "Go to FetchMessageBuffer.FindBodySection")

func (buf \*[FetchMessageBuffer](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageBuffer)) FindBodySection(section \*imap.FetchItemBodySection) \[\][byte](https://pkg.go.dev/builtin#byte)

FindBodySection returns the contents of a requested body section.

If the body section is not found, nil is returned.

type FetchMessageData struct {
	SeqNum [uint32](https://pkg.go.dev/builtin#uint32)
	
}

FetchMessageData contains a message's FETCH data.

Collect accumulates message data into a struct.

This method will read and store message contents in memory. This is acceptable when the message contents have a reasonable size, but may not be suitable when fetching e.g. attachments.

func (data \*[FetchMessageData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageData)) Next() [FetchItemData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchItemData)

Next advances to the next data item for this message.

If there is one or more data items left, the next item is returned. Otherwise nil is returned.

#### type [GetACLCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/acl.go#L55) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetACLCommand "Go to GetACLCommand")

type GetACLCommand struct {
	
}

GetACLCommand is a GETACL command.

#### func (\*GetACLCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/acl.go#L60) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetACLCommand.Wait "Go to GetACLCommand.Wait")

type GetACLData struct {
	Mailbox [string](https://pkg.go.dev/builtin#string)
	Rights  map\[imap.RightsIdentifier\]imap.RightSet
}

GetACLData is the data returned by the GETACL command.

#### type [GetMetadataCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/metadata.go#L134) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataCommand "Go to GetMetadataCommand")

type GetMetadataCommand struct {
	
}

GetMetadataCommand is a GETMETADATA command.

#### func (\*GetMetadataCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/metadata.go#L140) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataCommand.Wait "Go to GetMetadataCommand.Wait")

GetMetadataData is the data returned by the GETMETADATA command.

type GetMetadataDepth [int](https://pkg.go.dev/builtin#int)

const (
	GetMetadataDepthZero     [GetMetadataDepth](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataDepth) = 0
	GetMetadataDepthOne      [GetMetadataDepth](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataDepth) = 1
	GetMetadataDepthInfinity [GetMetadataDepth](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataDepth) = -1
)

type GetMetadataOptions struct {
	MaxSize \*[uint32](https://pkg.go.dev/builtin#uint32)
	Depth   [GetMetadataDepth](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetMetadataDepth)
}

GetMetadataOptions contains options for the GETMETADATA command.

#### type [GetQuotaCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/quota.go#L104) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaCommand "Go to GetQuotaCommand")

type GetQuotaCommand struct {
	
}

GetQuotaCommand is a GETQUOTA command.

#### func (\*GetQuotaCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/quota.go#L110) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaCommand.Wait "Go to GetQuotaCommand.Wait")

#### type [GetQuotaRootCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/quota.go#L118) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaRootCommand "Go to GetQuotaRootCommand")

type GetQuotaRootCommand struct {
	
}

GetQuotaRootCommand is a GETQUOTAROOT command.

#### func (\*GetQuotaRootCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/quota.go#L125) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#GetQuotaRootCommand.Wait "Go to GetQuotaRootCommand.Wait")

#### type [IDCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/id.go#L156) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IDCommand "Go to IDCommand")

type IDCommand struct {
	
}

#### func (\*IDCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/id.go#L161) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IDCommand.Wait "Go to IDCommand.Wait")

func (r \*[IDCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IDCommand)) Wait() (\*imap.IDData, [error](https://pkg.go.dev/builtin#error))

#### type [IdleCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/idle.go#L41) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IdleCommand "Go to IdleCommand")

type IdleCommand struct {
	
}

IdleCommand is an IDLE command.

Initially, the IDLE command is running. The server may send unilateral data. The client cannot send any command while IDLE is running.

Close must be called to stop the IDLE command.

#### func (\*IdleCommand) [Close](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/idle.go#L89) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IdleCommand.Close "Go to IdleCommand.Close")

Close stops the IDLE command.

This method blocks until the command to stop IDLE is written, but doesn't wait for the server to respond. Callers can use Wait for this purpose.

#### func (\*IdleCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/idle.go#L99) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#IdleCommand.Wait "Go to IdleCommand.Wait")

Wait blocks until the IDLE command has completed.

#### type [ListCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/list.go#L126) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand "Go to ListCommand")

type ListCommand struct {
	
}

ListCommand is a LIST command.

#### func (\*ListCommand) [Close](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/list.go#L146) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand.Close "Go to ListCommand.Close")

Close releases the command.

Calling Close unblocks the IMAP client decoder and lets it read the next responses. Next will always return nil after Close.

#### func (\*ListCommand) [Collect](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/list.go#L156) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand.Collect "Go to ListCommand.Collect")

func (cmd \*[ListCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand)) Collect() (\[\]\*imap.ListData, [error](https://pkg.go.dev/builtin#error))

Collect accumulates mailboxes into a list.

This is equivalent to calling Next repeatedly and then Close.

#### func (\*ListCommand) [Next](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/list.go#L138) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand.Next "Go to ListCommand.Next")

func (cmd \*[ListCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ListCommand)) Next() \*imap.ListData

Next advances to the next mailbox.

On success, the mailbox LIST data is returned. On error or if there are no more mailboxes, nil is returned.

#### type [MoveCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/move.go#L42) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MoveCommand "Go to MoveCommand")

type MoveCommand struct {
	
}

MoveCommand is a MOVE command.

#### func (\*MoveCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/move.go#L51) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MoveCommand.Wait "Go to MoveCommand.Wait")

type MoveData struct {
	
	UIDValidity [uint32](https://pkg.go.dev/builtin#uint32)
	SourceUIDs  imap.NumSet
	DestUIDs    imap.NumSet
}

MoveData contains the data returned by a MOVE command.

#### type [MyRightsCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/acl.go#L87) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MyRightsCommand "Go to MyRightsCommand")

type MyRightsCommand struct {
	
}

MyRightsCommand is a MYRIGHTS command.

#### func (\*MyRightsCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/acl.go#L92) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#MyRightsCommand.Wait "Go to MyRightsCommand.Wait")

type MyRightsData struct {
	Mailbox [string](https://pkg.go.dev/builtin#string)
	Rights  imap.RightSet
}

MyRightsData is the data returned by the MYRIGHTS command.

#### type [NamespaceCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/namespace.go#L31) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#NamespaceCommand "Go to NamespaceCommand")

type NamespaceCommand struct {
	
}

NamespaceCommand is a NAMESPACE command.

#### func (\*NamespaceCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/namespace.go#L36) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#NamespaceCommand.Wait "Go to NamespaceCommand.Wait")

func (cmd \*[NamespaceCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#NamespaceCommand)) Wait() (\*imap.NamespaceData, [error](https://pkg.go.dev/builtin#error))

Options contains options for Client.

type QuotaData struct {
	Root      [string](https://pkg.go.dev/builtin#string)
	Resources map\[imap.QuotaResourceType\][QuotaResourceData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#QuotaResourceData)
}

QuotaData is the data returned by a QUOTA response.

type QuotaResourceData struct {
	Usage [int64](https://pkg.go.dev/builtin#int64)
	Limit [int64](https://pkg.go.dev/builtin#int64)
}

QuotaResourceData contains the usage and limit for a quota resource.

#### type [SearchCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/search.go#L150) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SearchCommand "Go to SearchCommand")

type SearchCommand struct {
	
}

SearchCommand is a SEARCH command.

#### func (\*SearchCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/search.go#L155) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SearchCommand.Wait "Go to SearchCommand.Wait")

func (cmd \*[SearchCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SearchCommand)) Wait() (\*imap.SearchData, [error](https://pkg.go.dev/builtin#error))

#### type [SelectCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/select.go#L88) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectCommand "Go to SelectCommand")

type SelectCommand struct {
	
}

SelectCommand is a SELECT command.

#### func (\*SelectCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/select.go#L94) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectCommand.Wait "Go to SelectCommand.Wait")

func (cmd \*[SelectCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SelectCommand)) Wait() (\*imap.SelectData, [error](https://pkg.go.dev/builtin#error))

type SelectedMailbox struct {
	Name           [string](https://pkg.go.dev/builtin#string)
	NumMessages    [uint32](https://pkg.go.dev/builtin#uint32)
	Flags          \[\]imap.Flag
	PermanentFlags \[\]imap.Flag
}

SelectedMailbox contains metadata for the currently selected mailbox.

#### type [SetACLCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/acl.go#L35) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SetACLCommand "Go to SetACLCommand")

type SetACLCommand struct {
	
}

SetACLCommand is a SETACL command.

#### func (\*SetACLCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/acl.go#L39) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SetACLCommand.Wait "Go to SetACLCommand.Wait")

#### type [SortCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/sort.go#L76) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCommand "Go to SortCommand")

type SortCommand struct {
	
}

SortCommand is a SORT command.

#### func (\*SortCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/sort.go#L81) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCommand.Wait "Go to SortCommand.Wait")

type SortCriterion struct {
	Key     [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey)
	Reverse [bool](https://pkg.go.dev/builtin#bool)
}

const (
	SortKeyArrival [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey) = "ARRIVAL"
	SortKeyCc      [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey) = "CC"
	SortKeyDate    [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey) = "DATE"
	SortKeyFrom    [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey) = "FROM"
	SortKeySize    [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey) = "SIZE"
	SortKeySubject [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey) = "SUBJECT"
	SortKeyTo      [SortKey](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortKey) = "TO"
)

type SortOptions struct {
	SearchCriteria \*imap.SearchCriteria
	SortCriteria   \[\][SortCriterion](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#SortCriterion)
}

SortOptions contains options for the SORT command.

#### type [StatusCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/status.go#L81) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#StatusCommand "Go to StatusCommand")

type StatusCommand struct {
	
}

StatusCommand is a STATUS command.

#### func (\*StatusCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/status.go#L87) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#StatusCommand.Wait "Go to StatusCommand.Wait")

func (cmd \*[StatusCommand](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#StatusCommand)) Wait() (\*imap.StatusData, [error](https://pkg.go.dev/builtin#error))

#### type [ThreadCommand](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/thread.go#L54) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadCommand "Go to ThreadCommand")

type ThreadCommand struct {
	
}

ThreadCommand is a THREAD command.

#### func (\*ThreadCommand) [Wait](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/thread.go#L59) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadCommand.Wait "Go to ThreadCommand.Wait")

type ThreadData struct {
	Chain      \[\][uint32](https://pkg.go.dev/builtin#uint32)
	SubThreads \[\][ThreadData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#ThreadData)
}

type ThreadOptions struct {
	Algorithm      imap.ThreadAlgorithm
	SearchCriteria \*imap.SearchCriteria
}

ThreadOptions contains options for the THREAD command.

#### type [UnilateralDataHandler](https://github.com/emersion/go-imap/blob/v2.0.0-beta.5/imapclient/client.go#L1164) [¶](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#UnilateralDataHandler "Go to UnilateralDataHandler")

type UnilateralDataHandler struct {
	Expunge func(seqNum [uint32](https://pkg.go.dev/builtin#uint32))
	Mailbox func(data \*[UnilateralDataMailbox](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#UnilateralDataMailbox))
	Fetch   func(msg \*[FetchMessageData](https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient#FetchMessageData))
	
	Metadata func(mailbox [string](https://pkg.go.dev/builtin#string), entries \[\][string](https://pkg.go.dev/builtin#string))
}

UnilateralDataHandler handles unilateral data.

The handler will block the client while running. If the caller intends to perform slow operations, a buffered channel and a separate goroutine should be used.

The handler will be invoked in an arbitrary goroutine.

See Options.UnilateralDataHandler.

type UnilateralDataMailbox struct {
	NumMessages    \*[uint32](https://pkg.go.dev/builtin#uint32)
	Flags          \[\]imap.Flag
	PermanentFlags \[\]imap.Flag
}

UnilateralDataMailbox describes a mailbox status update.

If a field is nil, it hasn't changed.