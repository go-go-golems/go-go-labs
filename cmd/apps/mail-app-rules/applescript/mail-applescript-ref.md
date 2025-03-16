I'll convert this AppleScript reference documentation to clean Markdown format:

# Standard Suite

## open v : Open a document.
- **open** file or list of file : The file(s) to be opened.
  → **document** or list of **document** : The opened document(s).

## save options *enum*
- **yes** : Save the file.
- **no** : Do not save the file.
- **ask** : Ask the user whether or not to save the file.

## close v : Close a document.
- **close** specifier : the document(s) or window(s) to close.
  - [**saving** yes/no/ask] : Should changes be saved before closing?
  - [**saving in** file] : The file in which to save the document, if so.

## save v : Save a document.
- **save** specifier : The document(s) or window(s) to save.
  - [**in** file] : The file in which to save the document.
  - [**as** native format] : The file format to use.

## printing error handling *enum*
- **standard** : Standard PostScript error handling
- **detailed** : print a detailed report of PostScript errors

## print settings *n*
### PROPERTIES
- **copies** (integer) : the number of copies of a document to be printed
- **collating** (boolean) : Should printed copies be collated?
- **starting page** (integer) : the first page of the document to be printed
- **ending page** (integer) : the last page of the document to be printed
- **pages across** (integer) : number of logical pages laid across a physical page
- **pages down** (integer) : number of logical pages laid out down a physical page
- **requested print time** (date) : the time at which the desktop printer should print the document
- **error handling** (standard/detailed) : how errors are handled
- **fax number** (text) : for fax printer
- **target printer** (text) : for target printer

## print v : Print a document.
- **print** specifier : The document(s), window(s), or object(s) to be printed.
  - [**with properties** print settings] : The print settings to use.
  - [**print dialog** boolean] : Should the application show the print dialog?

## quit v : Quit the application.
- **quit**
  - [**saving** yes/no/ask] : Should changes be saved before quitting?

## count v : Return the number of elements of a particular class within an object.
- **count** specifier : The objects to be counted.
  → **integer** : The count.

## exists v : Verify that an object exists.
- **exists** any : The object(s) to check.
  → **boolean** : Did the object(s) exist?

## make v : Create a new object.
- **make**
  - **new** type : The class of the new object.
  - [**at** location specifier] : The location at which to insert the object.
  - [**with data** any] : The initial contents of the object.
  - [**with properties** record] : The initial values for properties of the object.
  → **specifier** : The new object.

## application *n* [see also Mail] : The application's top-level scripting object.
### ELEMENTS
- contains **documents**, **windows**.
### PROPERTIES
- **name** (text, r/o) : The name of the application.
- **frontmost** (boolean, r/o) : Is this the active application?
- **version** (text, r/o) : The version number of the application.
### RESPONDS TO
- **open**, **print**, **quit**.

## document *n* : A document.
### ELEMENTS
- contained by **application**.
### PROPERTIES
- **name** (text, r/o) : Its name.
- **modified** (boolean, r/o) : Has it been modified since the last save?
- **file** (file, r/o) : Its location on disk, if it has one.
### RESPONDS TO
- **close**, **print**, **save**.

## window *n* : A window.
### ELEMENTS
- contained by **application**.
### PROPERTIES
- **name** (text, r/o) : The title of the window.
- **id** (integer, r/o) : The unique identifier of the window.
- **index** (integer) : The index of the window, ordered front to back.
- **bounds** (rectangle) : The bounding rectangle of the window.
- **closeable** (boolean, r/o) : Does the window have a close button?
- **minimizable** (boolean, r/o) : Does the window have a minimize button?
- **minimized** (boolean) : Is the window minimized right now?
- **resizable** (boolean, r/o) : Can the window be resized?
- **visible** (boolean) : Is the window visible right now?
- **zoomable** (boolean, r/o) : Does the window have a zoom button?
- **zoomed** (boolean) : Is the window zoomed right now?
- **document** (**document**, r/o) : The document whose contents are displayed in the window.
### RESPONDS TO
- **close**, **print**, **save**.

## duplicate v : Copy an object.
- **duplicate** specifier : The object(s) to copy.
  - [**to** location specifier] : The location for the new copy or copies.
  - [**with properties** record] : Properties to set in the new copy or copies right away.

## move v : Move an object to a new location.
- **move** specifier : The object(s) to move.
  - **to** location specifier : The new location for the object(s).

---

I'll continue converting the AppleScript documentation to clean Markdown format:

# Text Suite

## RGB color *n*

## rich text *n, pl* rich text : Rich (styled) text
### ELEMENTS
- contains **paragraphs**, **words**, **characters**, **attribute runs**, **attachments**.
### PROPERTIES
- **color** (**RGB color**) : The color of the first character.
- **font** (text) : The name of the font of the first character.
- **size** (number) : The size in points of the first character.

## attachment *n* [inh. **rich text**] : Represents an inline text attachment. This class is used mainly for make commands
### ELEMENTS
- contained by **rich text**, **paragraphs**, **words**, **characters**, **attribute runs**.
### PROPERTIES
- **file name** (file) : The file for the attachment

## paragraph *n* : This subdivides the text into paragraphs.
### ELEMENTS
- contains **words**, **characters**, **attribute runs**, **attachments**; contained by **rich text**, **attribute runs**.
### PROPERTIES
- **color** (**RGB color**) : The color of the first character.
- **font** (text) : The name of the font of the first character.
- **size** (number) : The size in points of the first character.

## word *n* : This subdivides the text into words.
### ELEMENTS
- contains **characters**, **attribute runs**, **attachments**; contained by **rich text**, **paragraphs**, **attribute runs**.
### PROPERTIES
- **color** (**RGB color**) : The color of the first character.
- **font** (text) : The name of the font of the first character.
- **size** (number) : The size in points of the first character.

## character *n* : This subdivides the text into characters.
### ELEMENTS
- contains **attribute runs**, **attachments**; contained by **rich text**, **paragraphs**, **words**, **attribute runs**.
### PROPERTIES
- **color** (**RGB color**) : The color of the character.
- **font** (text) : The name of the font of the character.
- **size** (number) : The size in points of the character.

## attribute run *n* : This subdivides the text into chunks that all have the same attributes.
### ELEMENTS
- contains **paragraphs**, **words**, **characters**, **attachments**; contained by **rich text**, **paragraphs**, **words**, **characters**.
### PROPERTIES
- **color** (**RGB color**) : The color of the first character.
- **font** (text) : The name of the font of the first character.
- **size** (number) : The size in points of the first character.

---

# Mail

Classes and commands for the Mail application

## check for new mail v : Triggers a check for email.
- **check for new mail**
  - [**for** account] : Specify the account that you wish to check for mail

## extract name from v : Command to get the full name out of a fully specified email address. E.g. Calling this with "John Doe <jdoe@example.com>" as the direct object would return "John Doe"
- **extract name from** text : fully formatted email address
  → **text** : the full name

## extract address from v : Command to get just the email address of a fully specified email address. E.g. Calling this with "John Doe <jdoe@example.com>" as the direct object would return "jdoe@example.com"
- **extract address from** text : fully formatted email address
  → **text** : the email address

## forward v : Creates a forwarded message.
- **forward** message : the message to forward
  - [**opening window** boolean] : Whether the window for the forwarded message is shown. Default is to not show the window.
  → **outgoing message** : The message to be forwarded

## GetURL v : Opens a mailto URL.
- **GetURL** text : the mailto URL

## import Mail mailbox v : Imports a mailbox created by Mail.
- **import Mail mailbox**
  - **at** file : the mailbox or folder of mailboxes to import

## mailto v : Opens a mailto URL.
- **mailto** text : the mailto URL

## perform mail action with messages v : Script handler invoked by rules and menus that execute AppleScripts. The direct parameter of this handler is a list of messages being acted upon.
- **perform mail action with messages** list of message : the message being acted upon
  - [**in mailboxes** mailbox] : If the script is being executed by the user selecting an item in the scripts menu, this argument will specify the mailboxes that are currently selected. Otherwise it will not be specified.
  - [**for rule** rule] : If the script is being executed by a rule action, this argument will be the rule being invoked. Otherwise it will not be specified.

## redirect v : Creates a redirected message.
- **redirect** message : the message to redirect
  - [**opening window** boolean] : Whether the window for the redirected message is shown. Default is to not show the window.
  → **outgoing message** : The redirected message

## reply v : Creates a reply message.
- **reply** message : the message to reply to
  - [**opening window** boolean] : Whether the window for the reply message is shown. Default is to not show the window.
  - [**reply to all** boolean] : Whether to reply to all recipients. Default is to reply to the sender only.
  → **outgoing message** : the reply message

## send v : Sends a message.
- **send** outgoing message : the message to send
  → **boolean** : True if sending was successful, false if not

## synchronize v : Command to trigger synchronizing of an IMAP account with the server.
- **synchronize**
  - **with** account : The account to synchronize

## outgoing message n : A new email message
### ELEMENTS
- contains **bcc recipients**, **cc recipients**, **recipients**, **to recipients**; contained by **application**.
### PROPERTIES
- **sender** (text) : The sender of the message
- **subject** (text) : The subject of the message
- **content** (rich text) : The contents of the message
- **visible** (boolean) : Controls whether the message window is shown on the screen. The default is false
- **message signature** (signature or missing value) : The signature of the message
- **id** (integer, r/o) : The unique identifier of the message
### RESPONDS TO
- **save**, **close**, **send**.

## application n [see also Standard Suite] : Mail's top level scripting object.
### ELEMENTS
- contains **accounts**, **pop accounts**, **imap accounts**, **iCloud accounts**, **smtp servers**, **outgoing messages**, **mailboxes**, **message viewers**, **rules**, **signatures**.
### PROPERTIES
- **always bcc myself** (boolean) : Indicates whether you will be included in the Bcc: field of messages which you are composing
- **always cc myself** (boolean) : Indicates whether you will be included in the Cc: field of messages which you are composing
- **selection** (list of message, r/o) : List of messages that the user has selected
- **application version** (text, r/o) : The Mail version number in the format "3.0 (752/752.2)"
- **fetch interval** (integer) : The fetch interval (in minutes) between automatic fetches of new mail. -1 means to use an automatically determined interval
- **background activity count** (integer, r/o) : Number of background activities currently running in Mail, according to the Activity window
- **choose signature when composing** (boolean) : Indicates whether user can choose a signature directly in a new compose window
- **color quoted text** (boolean) : Indicates whether quoted text should be colored
- **default message format** (DefaultMessageFormat) : Default message format for messages which you are composing
- **default account** (account, r/o) : Default account to use for new outgoing messages
- **download remote images** (boolean) : Indicates whether images and attachments in HTML messages should be downloaded and displayed
- **drafts mailbox** (mailbox, r/o) : The top level Drafts mailbox
- **expand urls address fields** (boolean) : Indicates whether addresses will be expanded into a person with a prepend into the address fields of a new compose window
- **fixed width font** (text) : Font for plain text messages, only used if 'use fixed width font' is set to true
- **fixed width font size** (real) : Font size for plain text messages, only used if 'use fixed width font' is set to true
- **include all original message text** (boolean) : Indicates whether all of the original message will be quoted or only the text you have selected (if any)
- **quote original message** (boolean) : Indicates whether the text of the original message will be included in replies
- **check spelling while typing** (boolean) : Indicates whether spelling will be checked automatically in messages being composed
- **junk mailbox** (mailbox, r/o) : The top level Junk mailbox
- **level one quoting color** (blue/green/orange/other/purple/red/yellow) : Color for quoted text with one level of indentation
- **level two quoting color** (blue/green/orange/other/purple/red/yellow) : Color for quoted text with two levels of indentation
- **level three quoting color** (blue/green/orange/other/purple/red/yellow) : Color for quoted text with three levels of indentation
- **message font** (text) : Font for messages (proportional font)
- **message font size** (real) : Font size for messages with proportional font
- **message list font** (text) : Font for message list
- **message list font size** (real) : Font size for message list
- **new mail sound** (text) : Name of new mail sound or 'None' if no sound is selected
- **outbox mailbox** (mailbox, r/o) : The top level Out mailbox
- **should play other mail sounds** (boolean) : Indicates whether sounds will be played for various things such as when a messages is sent or if no mail is found when manually checking for new mail or if there is a fetch error
- **same reply format** (boolean) : Indicates whether replies will be in the same text format as the message to which you are replying
- **selected signature** (text) : Name of current selected signature (or 'randomly', 'sequentially', or 'none')
- **sent mailbox** (mailbox, r/o) : The top level Sent mailbox
- **fetches automatically** (boolean) : Indicates whether mail will automatically be fetched at a specific interval
- **highlight selected conversation** (boolean) : Indicates whether messages in conversations should be highlighted in the Mail viewer window when not grouped
- **trash mailbox** (mailbox, r/o) : The top level Trash mailbox
- **use address completion** (boolean) : This always returns true, and setting it doesn't do anything (deprecated)
- **use fixed width font** (boolean) : Should fixed-width font be used for plain text messages?
- **primary email** (text, r/o) : The user's primary email address
### RESPONDS TO
- **check for new mail**, **import Mail mailbox**, **synchronize**.

## message viewer n : Represents the object responsible for managing a viewer window
### ELEMENTS
- contains **messages**; contained by **application**.
### PROPERTIES
- **drafts mailbox** (mailbox, r/o) : The top level Drafts mailbox
- **inbox** (mailbox, r/o) : The top level In mailbox
- **junk mailbox** (mailbox, r/o) : The top level Junk mailbox
- **outbox** (mailbox, r/o) : The top level Out mailbox
- **sent mailbox** (mailbox, r/o) : The top level Sent mailbox
- **trash mailbox** (mailbox, r/o) : The top level Trash mailbox
- **sort column** (attachments column/message color/date received column/date sent column/flags column/from column/mailbox column/message status column/number column/size column/subject column/to column/date last saved column)
- **last sort column** : The column that is currently sorted in the viewer
- **sorted ascending** (boolean) : Whether the viewer is sorted ascending or not
- **mailbox list visible** (boolean) : Controls whether the list of mailboxes is visible or not
- **preview pane is visible** (boolean) : Controls whether the preview pane of the message viewer window is visible or not
- **visible columns** (list of attachments column/message color/date received column/date sent column/flags column/from column/mailbox column/message status column/number column/size column/subject column/to column/date last saved column) : List of columns that are visible. The subject column and the message status column will always be visible
- **id** (integer, r/o) : The unique identifier of the message viewer
- **visible messages** (list of message) : List of messages currently being displayed in the viewer
- **selected messages** (list of message) : List of messages currently selected
- **selected mailboxes** (list of mailbox) : List of mailboxes currently selected in the list of mailboxes
- **window** (window) : The window for the message viewer

## signature n : Email signatures
### ELEMENTS
- contained by **application**.
### PROPERTIES
- **content** (text) : Contents of email signature. If there is a version with fonts and/or styles, that will be returned over the plain text version
- **name** (text) : Name of the signature

## saveable file format enum
- **native format** : Native format

## DefaultMessageFormat enum
- **plain format** : Plain Text
- **rich format** : Rich Text

## QuotingColor enum
- **blue** : Blue
- **green** : Green
- **orange** : Orange
- **other** : Other
- **purple** : Purple
- **red** : Red
- **yellow** : Yellow

## ViewerColumns enum
- **attachments column** : Column containing the number of attachments a message contains
- **message color** : Used to indicate sorting should be done by color
- **date received column** : Column containing the date a message was received
- **date sent column** : Column containing the date a message was sent
- **flags column** : Column containing the flags of a message
- **from column** : Column containing the sender's name
- **mailbox column** : Column indicating the mailbox or account a message is in
- **message status column** : Column indicating a message status (read, unread, replied to, forwarded, etc)
- **number column** : Column containing the number of a message in a mailbox
- **size column** : Column containing the size of a message
- **subject column** : Column containing the subject of a message
- **to column** : Column containing the recipients of a message
- **date last saved column** : Column containing the date a draft message was saved

---

# Mail Framework

Classes and commands for the Mail framework

## message n : An email message
### ELEMENTS
- contains **bcc recipients**, **cc recipients**, **recipients**, **to recipients**, **headers**, **mail attachments**; contained by **message viewers**, **mailboxes**.
### PROPERTIES
- **id** (integer, r/o) : The unique identifier of the message.
- **all headers** (text, r/o) : All the headers of the message
- **background color** (blue/gray/green/none/orange/other/purple/red/yellow) : The background color of the message
- **mailbox** (mailbox) : The mailbox in which this message is filed
- **content** (text, r/o) : Contents of an email message
- **date received** (date, r/o) : The date a message was received
- **date sent** (date, r/o) : The date a message was sent
- **deleted status** (boolean) : Indicates whether the message is deleted or not
- **flagged status** (boolean) : Indicates whether the message is flagged or not
- **flag index** (integer) : The flag on the message, or -1 if the message is not flagged
- **junk mail status** (boolean) : Indicates whether the message has been marked junk or evaluated to be junk by the junk mail filter.
- **read status** (boolean) : Indicates whether the message is read or not
- **message size** (integer, r/o) : Size in bytes of a message
- **source** (text, r/o) : Raw source of the message
- **reply to** (text, r/o) : The address that replies should be sent to
- **message status** (text, r/o) : The status of a message of a message
- **sender** (text, r/o) : The sender of the message
- **subject** (text, r/o) : The subject of the message
- **was forwarded** (boolean, r/o) : Indicates whether the message was forwarded or not
- **was redirected** (boolean, r/o) : Indicates whether the message was redirected or not
- **was replied to** (boolean, r/o) : Indicates whether the message was replied to or not
### RESPONDS TO
- **open**, **bounce**, **forward**, **redirect**, **reply**.

## account n : A Mail account for receiving messages (POP/IMAP). To create a new receiving account, use the 'pop account', 'imap account', and 'iCloud account' objects
### ELEMENTS
- contains **mailboxes**; contained by **application**.
### PROPERTIES
- **delivery account** (smtp server or missing value) : The delivery account used when sending mail from this account
- **name** (text) : The name of an account
- **id** (text, r/o) : The unique identifier of the account
- **password** (text) : Password for this account. Can be set, but not read via scripting
- **authentication** (password/apop/kerberos 5/ntlm/md5/external/Apple token/none) : Preferred authentication scheme for account
- **account type** (pop/smtp/imap/iCloud, r/o) : The type of an account
- **email addresses** (list of text) : The list of email addresses configured for an account
- **full name** (text) : User name for the account
- **empty junk messages frequency** (integer) : Number of days before junk messages are deleted (0 = delete on quit, -1 = never delete)
- **empty trash frequency** (integer) : Number of days before messages in the trash are permanently deleted (0 = delete on quit, -1 = never delete)
- **empty junk messages on quit** (boolean) : Indicates whether the messages in the junk messages mailboxes will be deleted on quit
- **empty trash on quit** (boolean) : Indicates whether the messages in deleted messages mailboxes will be permanently deleted on quit
- **enabled** (boolean) : Indicates whether the account is enabled or not
- **user name** (text) : The user name used to connect to an account
- **account directory** (file, r/o) : The directory where the account stores things on disk
- **port** (integer) : The port used to connect to an account
- **server name** (text) : The host name used to connect to an account
- **move deleted messages to trash** (boolean) : Indicates whether messages that are deleted will be moved to the trash mailbox
- **uses ssl** (boolean) : Indicates whether SSL is enabled for this receiving account

## imap account n [inh. **account**] : An IMAP email account
### ELEMENTS
- contained by **application**.
### PROPERTIES
- **compact mailboxes when closing** (boolean) : Indicates whether an IMAP mailbox is automatically compacted when you quit Mail or switch to another mailbox
- **message caching** (all messages but omit attachments/all messages and their attachments) : Message caching setting for this account
- **store drafts on server** (boolean) : Indicates whether drafts will be stored on the IMAP server
- **store junk mail on server** (boolean) : Indicates whether junk mail will be stored on the IMAP server
- **store sent messages on server** (boolean) : Indicates whether sent messages will be stored on the IMAP server
- **store deleted messages on server** (boolean) : Indicates whether deleted messages will be stored on the IMAP server

## iCloud account n [inh. **imap account** > **account**] : An iCloud or MobileMe email account *syn* Mac account, MobileMe account
### ELEMENTS
- contained by **application**.

## pop account n [inh. **account**] : A POP email account
### ELEMENTS
- contained by **application**.
### PROPERTIES
- **big message warning size** (integer) : If message size (in bytes) is over this amount, Mail will prompt you asking whether you want to download the message (-1 = do not prompt)
- **delayed message deletion interval** (integer) : Number of days before messages that have been downloaded are deleted from the server (0 = delete immediately after downloading)
- **delete mail on server** (boolean) : Indicates whether POP account deletes messages on the server after downloading
- **delete messages when moved from inbox** (boolean) : Indicates whether messages will be deleted from the server when moved from your POP inbox

## smtp server n : An SMTP account (for sending email)
### ELEMENTS
- contained by **application**.
### PROPERTIES
- **name** (text) : The name of an account
- **password** (text) : Password for this account. Can be set, but not read via scripting
- **account type** (pop/smtp/imap/iCloud, r/o) : The type of an account
- **authentication** (password/apop/kerberos 5/ntlm/md5/external/Apple token/none) : Preferred authentication scheme for account
- **enabled** (boolean) : Indicates whether the account is enabled or not
- **port** (integer) : The port used to connect to an account
- **server name** (text) : The host name used to connect to an account
- **uses ssl** (boolean) : Indicates whether SSL is enabled for this receiving account

## mailbox n, pl mailboxes : A mailbox that holds messages
### ELEMENTS
- contains **mailboxes**, **messages**; contained by **application**, **accounts**, **mailboxes**.
### PROPERTIES
- **name** (text) : The name of a mailbox
- **unread count** (integer, r/o) : The number of unread messages in the mailbox
- **account** (account, r/o)
- **container** (mailbox, r/o)

## rule n : Class for message rules
### ELEMENTS
- contains **rule conditions**; contained by **application**.
### PROPERTIES
- **color message** (blue/gray/green/none/orange/other/purple/red/yellow) : If rule matches, apply this color
- **delete message** (boolean) : If rule matches, delete message
- **forward text** (text) : If rule matches, prepend this text to the forwarded message. Set to empty string to include no prepended text
- **forward message** (text) : If rule matches, forward message to this address, or multiple addresses, separated by commas. Set to empty string to disable this action
- **mark flagged** (boolean) : If rule matches, mark message as flagged
- **mark flag index** (integer) : If rule matches, mark message with the specified flag. Set to -1 to disable this action
- **mark read** (boolean) : If rule matches, mark message as read
- **play sound** (text) : If rule matches, play this sound (specify name of sound or path to sound)
- **redirect message** (text) : If rule matches, redirect message to this address, or multiple addresses, separated by commas. Set to empty string to disable this action
- **reply text** (text) : If rule matches, reply to message and prepend with this text. Set to empty string to disable this action
- **run script** (file or missing value) : If rule matches, run this compiled AppleScript. Set to empty string to disable this action
- **all conditions must be met** (boolean) : Indicates whether all conditions must be met for rule to execute
- **copy message** (mailbox) : If rule matches, copy to this mailbox
- **transfer message** (mailbox) : If rule matches, transfer to this mailbox
- **highlight text using color** (boolean) : Indicates whether the color will be used to highlight the text or background of a message in the message list
- **enabled** (boolean) : Indicates whether the rule is enabled
- **name** (text) : Name of rule
- **should copy message** (boolean) : Indicates whether the rule has a copy action
- **should move message** (boolean) : Indicates whether the rule has a move action
- **stop evaluating rules** (boolean) : If rule matches, stop rule evaluation for this message

## rule condition n : Class for conditions that can be attached to a single rule
### ELEMENTS
- contained by **rules**.
### PROPERTIES
- **expression** (text) : Rule expression field
- **header** (text) : Rule header key
- **qualifier** (begins with value/does contain value/does not contain value/ends with value/equal to value/less than value/greater than value/none) : Rule qualifier
- **rule type** (account/any recipient/cc header/matches every message/from header/header key/message content/message is junk mail/sender is in my contacts/sender is not in my contacts/sender is in my previous recipients/sender is not in my previous recipients/sender is VIP/subject header/to header/to or cc header/attachment type) : Rule type

## recipient n : An email recipient
### ELEMENTS
- contained by **outgoing messages**, **messages**.
### PROPERTIES
- **address** (text) : The recipients email address
- **name** (text) : The name used for display

## bcc recipient n [inh. **recipient**] : An email recipient in the Bcc: field

## cc recipient n [inh. **recipient**] : An email recipient in the Cc: field

## to recipient n [inh. **recipient**] : An email recipient in the To: field

## container n [inh. **mailbox**] : A mailbox that contains other mailboxes.

## header n : A header value for a message. E.g. To, Subject, From.
### ELEMENTS
- contained by **messages**.
### PROPERTIES
- **content** (text) : Contents of the header
- **name** (text) : Name of the header value

## mail attachment n : A file attached to a received message.
### ELEMENTS
- contained by **messages**.
### PROPERTIES
- **name** (text, r/o) : Name of the attachment.
- **mime type** (text, r/o) : MIME type of the attachment E.g. text/plain.
- **file size** (integer, r/o) : Approximate size in bytes.
- **downloaded** (boolean, r/o) : Indicates whether the attachment has been downloaded.
- **id** (text, r/o) : The unique identifier of the attachment.
### RESPONDS TO
- **save**.

## Authentication enum
- **password** : Clear text password
- **apop** : APOP
- **kerberos 5** : Kerberos V5 (GSSAPI)
- **ntlm** : NTLM
- **md5** : CRAM-MD5
- **external** : External authentication (TLS client certificate)
- **Apple token** : Apple token
- **none** : None

## HighlightColors enum
- **blue** : Blue
- **gray** : Gray
- **green** : Green
- **none** : None
- **orange** : Orange
- **other** : Other
- **purple** : Purple
- **red** : Red
- **yellow** : Yellow

## MessageCachingPolicy enum
- **all messages but omit attachments** : All messages but omit attachments
- **all messages and their attachments** : All messages and their attachments

## RuleQualifier enum
- **begins with value** : Begins with value
- **does contain value** : Contains value
- **does not contain value** : Does not contain value
- **ends with value** : Ends with value
- **equal to value** : Equal to value
- **less than value** : Less than value
- **greater than value** : Greater than value
- **none** : Indicates no qualifier is applicable

## RuleType enum
- **account** : Account
- **any recipient** : Any recipient
- **cc header** : Cc header
- **matches every message** : Every message
- **from header** : From header
- **header key** : An arbitrary header key
- **message content** : Message content
- **message is junk mail** : Message is junk mail
- **sender is in my contacts** : Sender is in my contacts
- **sender is in my previous recipients** : Sender is in my previous recipients
- **sender is member of group** : Sender is member of group
- **sender is not in my contacts** : Sender is not in my contacts
- **sender is not in my previous recipients** : Sender is not in my previous recipients
- **sender is not member of group** : Sender is not member of group
- **sender is VIP** : Sender is VIP
- **subject header** : Subject header
- **to header** : To header
- **to or cc header** : To or Cc header
- **attachment type** : Attachment Type

## TypeOfAccount enum
- **pop** : POP
- **smtp** : SMTP
- **imap** : IMAP
- **iCloud** : iCloud *syn* Mac