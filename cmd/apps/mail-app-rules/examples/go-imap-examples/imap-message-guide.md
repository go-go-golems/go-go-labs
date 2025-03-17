# Fetching and Processing Email Messages with go-imap and go-message

This guide explains how to fetch and process email messages over IMAP using the [go-imap](https://github.com/emersion/go-imap) and [go-message](https://github.com/emersion/go-message) libraries.

## Understanding IMAP

### What is IMAP?

IMAP (Internet Message Access Protocol) is a standard protocol for accessing and managing email messages stored on a mail server. Unlike POP3, which typically downloads messages to a local client, IMAP allows clients to access messages stored on the server, making it ideal for accessing email from multiple devices.

The current version of the protocol is IMAP4rev2, defined in [RFC 9051](https://www.rfc-editor.org/rfc/rfc9051.html).

### IMAP Concepts

#### Connection States

An IMAP connection can be in one of several states:

1. **Not Authenticated**: Initial state after connection
2. **Authenticated**: After successful login
3. **Selected**: After a mailbox is selected
4. **Logout**: When the connection is being terminated

#### Mailboxes

IMAP organizes messages into mailboxes (folders). Common mailboxes include:

- INBOX (required)
- Sent
- Drafts
- Trash
- Junk/Spam

#### Message Identification

Messages in IMAP can be identified by:

- **Sequence Numbers**: Temporary identifiers assigned to messages in a selected mailbox
- **UIDs**: Unique identifiers that remain stable across sessions (preferred for most operations)

### Message Fetching in IMAP

#### FETCH Command

The FETCH command is used to retrieve information about messages from the server. It can retrieve:

1. **Message Metadata**: Flags, size, internal date, etc.
2. **Message Structure**: BODYSTRUCTURE provides information about the message's MIME structure
3. **Message Content**: Full or partial message content

#### Understanding FETCH Data Items

When fetching messages, you specify which data items you want to retrieve:

- **ENVELOPE**: Contains header information (subject, from, to, date, etc.)
- **FLAGS**: Message flags (seen, answered, flagged, etc.)
- **INTERNALDATE**: When the message was received by the server
- **RFC822.SIZE**: Size of the message in bytes
- **UID**: Unique identifier for the message
- **BODYSTRUCTURE**: Detailed information about the message's MIME structure
- **BODY[]**: The full message content
- **BODY[section]**: A specific part of the message

#### BODYSTRUCTURE vs BODY

- **BODYSTRUCTURE**: Provides detailed information about the message's MIME structure without retrieving the actual content. This is useful for understanding the message's parts before deciding which parts to fetch.

- **BODY[]**: Retrieves the actual content of the message or specific parts.

#### Section Specifiers

When fetching message content, you can specify which part to retrieve:

- **BODY[]**: The entire message
- **BODY[HEADER]**: Just the message headers
- **BODY[TEXT]**: Just the message body
- **BODY[1]**: The first MIME part
- **BODY[1.2]**: The second part of the first part (for nested multipart messages)
- **BODY[HEADER.FIELDS (From To Subject)]**: Only specific header fields

## Using go-imap and go-message

### Connecting to an IMAP Server

```go
package main

import (
    "log"
    
    "github.com/emersion/go-imap/v2/imapclient"
)

func main() {
    // Connect to the server
    client, err := imapclient.DialTLS("imap.example.com:993", nil)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()
    
    // Login
    if err := client.Login("username", "password").Wait(); err != nil {
        log.Fatalf("Failed to login: %v", err)
    }
    
    // Select a mailbox
    _, err = client.Select("INBOX", nil).Wait()
    if err != nil {
        log.Fatalf("Failed to select INBOX: %v", err)
    }
    
    // Logout when done
    if err := client.Logout().Wait(); err != nil {
        log.Fatalf("Failed to logout: %v", err)
    }
}
```

### Fetching Message Metadata

To fetch basic information about messages:

```go
package main

import (
    "log"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
)

func main() {
    // ... connect and login as shown above
    
    // Fetch the last 10 messages
    seqSet, _ := imap.ParseSeqSet("1:10")
    fetchOptions := &imap.FetchOptions{
        Envelope: true,
        Flags: true,
        InternalDate: true,
        RFC822Size: true,
        UID: true,
    }
    
    messages, err := client.Fetch(seqSet, fetchOptions).Collect()
    if err != nil {
        log.Fatalf("Failed to fetch messages: %v", err)
    }
    
    // Process the messages
    for _, msg := range messages {
        log.Printf("UID: %d, Subject: %s, Date: %s, Size: %d bytes",
            msg.UID, msg.Envelope.Subject, msg.InternalDate, msg.RFC822Size)
        log.Printf("Flags: %v", msg.Flags)
    }
}
```

### Fetching Message Structure

To understand the structure of a message before fetching its content:

```go
package main

import (
    "log"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
)

func main() {
    // ... connect and login as shown above
    
    // Fetch the structure of a specific message
    uid := imap.UID(12345)
    fetchOptions := &imap.FetchOptions{
        BodyStructure: &imap.FetchItemBodyStructure{
            Extended: true, // Get detailed structure
        },
    }
    
    messages, err := client.Fetch(imap.UIDSetNum(uid), fetchOptions).Collect()
    if err != nil {
        log.Fatalf("Failed to fetch message structure: %v", err)
    }
    
    if len(messages) == 0 {
        log.Fatalf("Message not found")
    }
    
    // Process the structure
    msg := messages[0]
    if msg.BodyStructure != nil {
        // Walk through the structure
        msg.BodyStructure.Walk(func(path []int, part imap.BodyStructure) bool {
            mediaType := part.MediaType()
            log.Printf("Part %v: %s", path, mediaType)
            
            // For single parts, we can get more details
            if singlePart, ok := part.(*imap.BodyStructureSinglePart); ok {
                log.Printf("  Type: %s/%s", singlePart.Type, singlePart.Subtype)
                log.Printf("  Encoding: %s", singlePart.Encoding)
                log.Printf("  Size: %d bytes", singlePart.Size)
                
                // Check for filename in Content-Disposition
                if singlePart.Extended != nil && singlePart.Extended.Disposition != nil {
                    if filename, ok := singlePart.Extended.Disposition.Params["filename"]; ok {
                        log.Printf("  Filename: %s", filename)
                    }
                }
            }
            
            return true // Continue walking
        })
    }
}
```

### Fetching Message Content

To fetch the full content of a message:

```go
package main

import (
    "io"
    "log"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
    "github.com/emersion/go-message/mail"
)

func main() {
    // ... connect and login as shown above
    
    // Fetch the content of a specific message
    uid := imap.UID(12345)
    bodySection := &imap.FetchItemBodySection{} // Empty means fetch the entire message
    fetchOptions := &imap.FetchOptions{
        BodySection: []*imap.FetchItemBodySection{bodySection},
    }
    
    fetchCmd := client.Fetch(imap.UIDSetNum(uid), fetchOptions)
    defer fetchCmd.Close()
    
    // Get the first message
    msg := fetchCmd.Next()
    if msg == nil {
        log.Fatalf("Message not found")
    }
    
    // Find the body section in the response
    var bodySectionData imapclient.FetchItemDataBodySection
    var found bool
    
    for {
        item := msg.Next()
        if item == nil {
            break
        }
        
        if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
            bodySectionData = data
            found = true
            break
        }
    }
    
    if !found {
        log.Fatalf("Body section not found in response")
    }
    
    // Parse the message using go-message
    mr, err := mail.CreateReader(bodySectionData.Literal)
    if err != nil {
        log.Fatalf("Failed to create mail reader: %v", err)
    }
    
    // Process the message header
    header := mr.Header
    if date, err := header.Date(); err == nil {
        log.Printf("Date: %v", date)
    }
    if from, err := header.AddressList("From"); err == nil {
        log.Printf("From: %v", from)
    }
    if to, err := header.AddressList("To"); err == nil {
        log.Printf("To: %v", to)
    }
    if subject, err := header.Subject(); err == nil {
        log.Printf("Subject: %v", subject)
    }
    
    // Process each part of the message
    for {
        part, err := mr.NextPart()
        if err == io.EOF {
            break
        } else if err != nil {
            log.Fatalf("Failed to read message part: %v", err)
        }
        
        switch header := part.Header.(type) {
        case *mail.InlineHeader:
            // This is the message's text (can be plain-text or HTML)
            contentType, _, _ := header.ContentType()
            content, _ := io.ReadAll(part.Body)
            log.Printf("Text part (%s): %s", contentType, string(content))
            
        case *mail.AttachmentHeader:
            // This is an attachment
            filename, _ := header.Filename()
            contentType, _, _ := header.ContentType()
            
            // Read the attachment data
            data, err := io.ReadAll(part.Body)
            if err != nil {
                log.Printf("Failed to read attachment: %v", err)
                continue
            }
            
            log.Printf("Attachment: %s (%s), %d bytes", 
                filename, contentType, len(data))
            
            // Here you could save the attachment to a file
            // ioutil.WriteFile(filename, data, 0644)
        }
    }
}
```

### Fetching Specific Message Parts

To fetch only specific parts of a message:

```go
package main

import (
    "io"
    "log"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
)

func main() {
    // ... connect and login as shown above
    
    uid := imap.UID(12345)
    
    // First, fetch the structure to understand the parts
    structureOptions := &imap.FetchOptions{
        BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
    }
    
    structMsgs, err := client.Fetch(imap.UIDSetNum(uid), structureOptions).Collect()
    if err != nil || len(structMsgs) == 0 {
        log.Fatalf("Failed to fetch structure: %v", err)
    }
    
    // Now fetch specific parts based on the structure
    // For example, to fetch just the headers:
    headerSection := &imap.FetchItemBodySection{
        Specifier: imap.PartSpecifierHeader,
    }
    
    // To fetch a specific MIME part (e.g., the first attachment):
    // Assuming part 2 is an attachment based on the structure
    attachmentSection := &imap.FetchItemBodySection{
        Part: []int{2}, // This refers to the second MIME part
    }
    
    // Fetch both parts in one request
    fetchOptions := &imap.FetchOptions{
        BodySection: []*imap.FetchItemBodySection{
            headerSection,
            attachmentSection,
        },
    }
    
    messages, err := client.Fetch(imap.UIDSetNum(uid), fetchOptions).Collect()
    if err != nil || len(messages) == 0 {
        log.Fatalf("Failed to fetch parts: %v", err)
    }
    
    msg := messages[0]
    
    // Process the header
    headerData := msg.FindBodySection(headerSection)
    log.Printf("Headers:\n%s", string(headerData))
    
    // Process the attachment
    attachmentData := msg.FindBodySection(attachmentSection)
    log.Printf("Attachment data size: %d bytes", len(attachmentData))
}
```

### Streaming Large Messages

For large messages, it's better to stream the content rather than loading it all into memory:

```go
package main

import (
    "io"
    "log"
    "os"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
    "github.com/emersion/go-message/mail"
)

func main() {
    // ... connect and login as shown above
    
    uid := imap.UID(12345)
    bodySection := &imap.FetchItemBodySection{}
    fetchOptions := &imap.FetchOptions{
        BodySection: []*imap.FetchItemBodySection{bodySection},
    }
    
    fetchCmd := client.Fetch(imap.UIDSetNum(uid), fetchOptions)
    defer fetchCmd.Close()
    
    msg := fetchCmd.Next()
    if msg == nil {
        log.Fatalf("Message not found")
    }
    
    // Process items as they arrive
    for {
        item := msg.Next()
        if item == nil {
            break
        }
        
        if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
            // Create a mail reader that will stream the message
            mr, err := mail.CreateReader(data.Literal)
            if err != nil {
                log.Fatalf("Failed to create mail reader: %v", err)
            }
            
            // Process each part
            for {
                part, err := mr.NextPart()
                if err == io.EOF {
                    break
                } else if err != nil {
                    log.Fatalf("Failed to read part: %v", err)
                }
                
                switch header := part.Header.(type) {
                case *mail.AttachmentHeader:
                    filename, _ := header.Filename()
                    if filename == "" {
                        filename = "unknown_attachment"
                    }
                    
                    // Stream the attachment directly to a file
                    file, err := os.Create(filename)
                    if err != nil {
                        log.Printf("Failed to create file: %v", err)
                        continue
                    }
                    
                    n, err := io.Copy(file, part.Body)
                    file.Close()
                    if err != nil {
                        log.Printf("Failed to save attachment: %v", err)
                    } else {
                        log.Printf("Saved attachment %s (%d bytes)", filename, n)
                    }
                    
                default:
                    // For text parts, we might still want to read them into memory
                    data, _ := io.ReadAll(part.Body)
                    log.Printf("Part content (%d bytes): %s", len(data), string(data))
                }
            }
        }
    }
}
```

## Advanced Techniques

### Handling HTML and Plain Text Alternatives

Many emails contain both HTML and plain text versions. Here's how to handle them:

```go
package main

import (
    "io"
    "log"
    "strings"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
    "github.com/emersion/go-message/mail"
)

func main() {
    // ... fetch the message as shown above
    
    // After creating the mail reader:
    mr, _ := mail.CreateReader(bodySectionData.Literal)
    
    var plainText, htmlText string
    
    // Process each part
    for {
        part, err := mr.NextPart()
        if err == io.EOF {
            break
        } else if err != nil {
            log.Fatalf("Failed to read part: %v", err)
        }
        
        if header, ok := part.Header.(*mail.InlineHeader); ok {
            contentType, _, _ := header.ContentType()
            content, _ := io.ReadAll(part.Body)
            
            if strings.HasPrefix(contentType, "text/plain") {
                plainText = string(content)
            } else if strings.HasPrefix(contentType, "text/html") {
                htmlText = string(content)
            }
        }
    }
    
    // Prefer HTML if available, otherwise use plain text
    if htmlText != "" {
        log.Printf("HTML content: %s", htmlText)
    } else if plainText != "" {
        log.Printf("Plain text content: %s", plainText)
    } else {
        log.Printf("No text content found")
    }
}
```

### Handling Embedded Images

Many HTML emails contain embedded images (often called "inline attachments"):

```go
package main

import (
    "io"
    "log"
    "os"
    "strings"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
    "github.com/emersion/go-message/mail"
)

func main() {
    // ... fetch the message as shown above
    
    // After creating the mail reader:
    mr, _ := mail.CreateReader(bodySectionData.Literal)
    
    // Map to store Content-ID -> filename for embedded images
    embeddedImages := make(map[string]string)
    
    // Process each part
    for {
        part, err := mr.NextPart()
        if err == io.EOF {
            break
        } else if err != nil {
            log.Fatalf("Failed to read part: %v", err)
        }
        
        switch header := part.Header.(type) {
        case *mail.AttachmentHeader:
            // Check if this is an inline attachment
            contentID := header.Get("Content-ID")
            if contentID != "" {
                // Clean up the Content-ID (remove < and >)
                contentID = strings.Trim(contentID, "<>")
                
                // Get the filename or generate one
                filename, err := header.Filename()
                if err != nil || filename == "" {
                    filename = "inline_" + contentID
                }
                
                // Save the image
                file, err := os.Create(filename)
                if err != nil {
                    log.Printf("Failed to create file: %v", err)
                    continue
                }
                
                _, err = io.Copy(file, part.Body)
                file.Close()
                if err != nil {
                    log.Printf("Failed to save image: %v", err)
                    continue
                }
                
                // Store the mapping
                embeddedImages[contentID] = filename
                log.Printf("Saved embedded image %s with Content-ID %s", filename, contentID)
            }
        }
    }
    
    // Now you can process the HTML and replace cid: references with file paths
    // For example, in the HTML you might find: <img src="cid:image001.jpg@01D...">
    // You would replace this with: <img src="file:///path/to/saved/image001.jpg">
}
```

## Best Practices

1. **Use UIDs Instead of Sequence Numbers**: UIDs are stable across sessions, while sequence numbers can change.

2. **Fetch Only What You Need**: Don't fetch the entire message if you only need headers or specific parts.

3. **Stream Large Messages**: For large messages or attachments, stream the content rather than loading it all into memory.

4. **Handle Errors Gracefully**: IMAP servers can disconnect or timeout, so handle errors appropriately.

5. **Close Commands When Done**: Always close fetch commands with `fetchCmd.Close()` to release resources.

6. **Use BODYSTRUCTURE First**: For complex messages, fetch the structure first to understand what parts are available before fetching the content.

7. **Consider Using PEEK**: When fetching message content, use the PEEK option to avoid marking messages as read.

## Conclusion

The go-imap and go-message libraries provide powerful tools for working with email messages over IMAP. By understanding the IMAP protocol and using these libraries effectively, you can build robust email applications in Go.

For more information, refer to the official documentation:
- [go-imap documentation](https://pkg.go.dev/github.com/emersion/go-imap/v2)
- [go-message documentation](https://pkg.go.dev/github.com/emersion/go-message)
- [IMAP4rev2 specification (RFC 9051)](https://www.rfc-editor.org/rfc/rfc9051.html) 