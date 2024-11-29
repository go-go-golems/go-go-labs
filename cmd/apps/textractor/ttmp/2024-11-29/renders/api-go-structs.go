File: /home/manuel/go/pkg/mod/github.com/aws/aws-sdk-go@v1.55.5/service/textract/api.go

Struct Declarations:
  // You aren't authorized to perform the action. Use the Amazon Resource Name
// (ARN) of an authorized user or IAM role to perform the operation. 
  type AccessDeniedException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // An adapter selected for use when analyzing documents. Contains an adapter
// ID and a version number. Contains information on pages selected for analysis
// when analyzing documents asychronously. 
  type Adapter   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the adapter resource.
  	//
  	// AdapterId is a required field
  	AdapterId *string `min:"12" type:"string" required:"true"`
  
  	// Pages is a parameter that the user inputs to specify which pages to apply
  	// an adapter to. The following is a list of rules for using this parameter.
  	//
  	//    * If a page is not specified, it is set to ["1"] by default.
  	//
  	//    * The following characters are allowed in the parameter's string: 0 1
  	//    2 3 4 5 6 7 8 9 - *. No whitespace is allowed.
  	//
  	//    * When using * to indicate all pages, it must be the only element in the
  	//    list.
  	//
  	//    * You can use page intervals, such as ["1-3", "1-1", "4-*"]. Where * indicates
  	//    last page of document.
  	//
  	//    * Specified pages must be greater than 0 and less than or equal to the
  	//    number of pages in the document.
  	Pages []*string `min:"1" type:"list"`
  
  	// A string that identifies the version of the adapter.
  	//
  	// Version is a required field
  	Version *string `min:"1" type:"string" required:"true"`
  }

  // Contains information on the adapter, including the adapter ID, Name, Creation
// time, and feature types. 
  type AdapterOverview   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the adapter resource.
  	AdapterId *string `min:"12" type:"string"`
  
  	// A string naming the adapter resource.
  	AdapterName *string `min:"1" type:"string"`
  
  	// The date and time that the adapter was created.
  	CreationTime *time.Time `type:"timestamp"`
  
  	// The feature types that the adapter is operating on.
  	FeatureTypes []*string `type:"list" enum:"FeatureType"`
  }

  // The dataset configuration options for a given version of an adapter. Can
// include an Amazon S3 bucket if specified. 
  type AdapterVersionDatasetConfig   struct {
  	_ struct{} `type:"structure"`
  
  	// The S3 bucket name and file name that identifies the document.
  	//
  	// The AWS Region for the S3 bucket that contains the document must match the
  	// Region that you use for Amazon Textract operations.
  	//
  	// For Amazon Textract to process a file in an S3 bucket, the user must have
  	// permission to access the S3 bucket and file.
  	ManifestS3Object *S3Object `type:"structure"`
  }

  // Contains information on the metrics used to evalute the peformance of a given
// adapter version. Includes data for baseline model performance and individual
// adapter version perfromance. 
  type AdapterVersionEvaluationMetric   struct {
  	_ struct{} `type:"structure"`
  
  	// The F1 score, precision, and recall metrics for the baseline model.
  	AdapterVersion *EvaluationMetric `type:"structure"`
  
  	// The F1 score, precision, and recall metrics for the baseline model.
  	Baseline *EvaluationMetric `type:"structure"`
  
  	// Indicates the feature type being analyzed by a given adapter version.
  	FeatureType *string `type:"string" enum:"FeatureType"`
  }

  // Summary info for an adapter version. Contains information on the AdapterId,
// AdapterVersion, CreationTime, FeatureTypes, and Status. 
  type AdapterVersionOverview   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the adapter associated with a given adapter version.
  	AdapterId *string `min:"12" type:"string"`
  
  	// An identified for a given adapter version.
  	AdapterVersion *string `min:"1" type:"string"`
  
  	// The date and time that a given adapter version was created.
  	CreationTime *time.Time `type:"timestamp"`
  
  	// The feature types that the adapter version is operating on.
  	FeatureTypes []*string `type:"list" enum:"FeatureType"`
  
  	// Contains information on the status of a given adapter version.
  	Status *string `type:"string" enum:"AdapterVersionStatus"`
  
  	// A message explaining the status of a given adapter vesion.
  	StatusMessage *string `min:"1" type:"string"`
  }

  // Contains information about adapters used when analyzing a document, with
// each adapter specified using an AdapterId and version 
  type AdaptersConfig   struct {
  	_ struct{} `type:"structure"`
  
  	// A list of adapters to be used when analyzing the specified document.
  	//
  	// Adapters is a required field
  	Adapters []*Adapter `min:"1" type:"list" required:"true"`
  }

  
  type AnalyzeDocumentInput   struct {
  	_ struct{} `type:"structure"`
  
  	// Specifies the adapter to be used when analyzing a document.
  	AdaptersConfig *AdaptersConfig `type:"structure"`
  
  	// The input document as base64-encoded bytes or an Amazon S3 object. If you
  	// use the AWS CLI to call Amazon Textract operations, you can't pass image
  	// bytes. The document must be an image in JPEG, PNG, PDF, or TIFF format.
  	//
  	// If you're using an AWS SDK to call Amazon Textract, you might not need to
  	// base64-encode image bytes that are passed using the Bytes field.
  	//
  	// Document is a required field
  	Document *Document `type:"structure" required:"true"`
  
  	// A list of the types of analysis to perform. Add TABLES to the list to return
  	// information about the tables that are detected in the input document. Add
  	// FORMS to return detected form data. Add SIGNATURES to return the locations
  	// of detected signatures. Add LAYOUT to the list to return information about
  	// the layout of the document. All lines and words detected in the document
  	// are included in the response (including text that isn't related to the value
  	// of FeatureTypes).
  	//
  	// FeatureTypes is a required field
  	FeatureTypes []*string `type:"list" required:"true" enum:"FeatureType"`
  
  	// Sets the configuration for the human in the loop workflow for analyzing documents.
  	HumanLoopConfig *HumanLoopConfig `type:"structure"`
  
  	// Contains Queries and the alias for those Queries, as determined by the input.
  	QueriesConfig *QueriesConfig `type:"structure"`
  }

  
  type AnalyzeDocumentOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The version of the model used to analyze the document.
  	AnalyzeDocumentModelVersion *string `type:"string"`
  
  	// The items that are detected and analyzed by AnalyzeDocument.
  	Blocks []*Block `type:"list"`
  
  	// Metadata about the analyzed document. An example is the number of pages.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// Shows the results of the human in the loop evaluation.
  	HumanLoopActivationOutput *HumanLoopActivationOutput `type:"structure"`
  }

  
  type AnalyzeExpenseInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The input document, either as bytes or as an S3 object.
  	//
  	// You pass image bytes to an Amazon Textract API operation by using the Bytes
  	// property. For example, you would use the Bytes property to pass a document
  	// loaded from a local file system. Image bytes passed by using the Bytes property
  	// must be base64 encoded. Your code might not need to encode document file
  	// bytes if you're using an AWS SDK to call Amazon Textract API operations.
  	//
  	// You pass images stored in an S3 bucket to an Amazon Textract API operation
  	// by using the S3Object property. Documents stored in an S3 bucket don't need
  	// to be base64 encoded.
  	//
  	// The AWS Region for the S3 bucket that contains the S3 object must match the
  	// AWS Region that you use for Amazon Textract operations.
  	//
  	// If you use the AWS CLI to call Amazon Textract operations, passing image
  	// bytes using the Bytes property isn't supported. You must first upload the
  	// document to an Amazon S3 bucket, and then call the operation using the S3Object
  	// property.
  	//
  	// For Amazon Textract to process an S3 object, the user must have permission
  	// to access the S3 object.
  	//
  	// Document is a required field
  	Document *Document `type:"structure" required:"true"`
  }

  
  type AnalyzeExpenseOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// Information about the input document.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// The expenses detected by Amazon Textract.
  	ExpenseDocuments []*ExpenseDocument `type:"list"`
  }

  // Used to contain the information detected by an AnalyzeID operation. 
  type AnalyzeIDDetections   struct {
  	_ struct{} `type:"structure"`
  
  	// The confidence score of the detected text.
  	Confidence *float64 `type:"float"`
  
  	// Only returned for dates, returns the type of value detected and the date
  	// written in a more machine readable way.
  	NormalizedValue *NormalizedValue `type:"structure"`
  
  	// Text of either the normalized field or value associated with it.
  	//
  	// Text is a required field
  	Text *string `type:"string" required:"true"`
  }

  
  type AnalyzeIDInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The document being passed to AnalyzeID.
  	//
  	// DocumentPages is a required field
  	DocumentPages []*Document `min:"1" type:"list" required:"true"`
  }

  
  type AnalyzeIDOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The version of the AnalyzeIdentity API being used to process documents.
  	AnalyzeIDModelVersion *string `type:"string"`
  
  	// Information about the input document.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// The list of documents processed by AnalyzeID. Includes a number denoting
  	// their place in the list and the response structure for the document.
  	IdentityDocuments []*IdentityDocument `type:"list"`
  }

  // Amazon Textract isn't able to read the document. For more information on
// the document limits in Amazon Textract, see limits. 
  type BadDocumentException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // A Block represents items that are recognized in a document within a group
// of pixels close to each other. The information returned in a Block object
// depends on the type of operation. In text detection for documents (for example
// DetectDocumentText), you get information about the detected words and lines
// of text. In text analysis (for example AnalyzeDocument), you can also get
// information about the fields, tables, and selection elements that are detected
// in the document.
//
// An array of Block objects is returned by both synchronous and asynchronous
// operations. In synchronous operations, such as DetectDocumentText, the array
// of Block objects is the entire set of results. In asynchronous operations,
// such as GetDocumentAnalysis, the array is returned over one or more responses.
//
// For more information, see How Amazon Textract Works (https://docs.aws.amazon.com/textract/latest/dg/how-it-works.html). 
  type Block   struct {
  	_ struct{} `type:"structure"`
  
  	// The type of text item that's recognized. In operations for text detection,
  	// the following types are returned:
  	//
  	//    * PAGE - Contains a list of the LINE Block objects that are detected on
  	//    a document page.
  	//
  	//    * WORD - A word detected on a document page. A word is one or more ISO
  	//    basic Latin script characters that aren't separated by spaces.
  	//
  	//    * LINE - A string of tab-delimited, contiguous words that are detected
  	//    on a document page.
  	//
  	// In text analysis operations, the following types are returned:
  	//
  	//    * PAGE - Contains a list of child Block objects that are detected on a
  	//    document page.
  	//
  	//    * KEY_VALUE_SET - Stores the KEY and VALUE Block objects for linked text
  	//    that's detected on a document page. Use the EntityType field to determine
  	//    if a KEY_VALUE_SET object is a KEY Block object or a VALUE Block object.
  	//
  	//    * WORD - A word that's detected on a document page. A word is one or more
  	//    ISO basic Latin script characters that aren't separated by spaces.
  	//
  	//    * LINE - A string of tab-delimited, contiguous words that are detected
  	//    on a document page.
  	//
  	//    * TABLE - A table that's detected on a document page. A table is grid-based
  	//    information with two or more rows or columns, with a cell span of one
  	//    row and one column each.
  	//
  	//    * TABLE_TITLE - The title of a table. A title is typically a line of text
  	//    above or below a table, or embedded as the first row of a table.
  	//
  	//    * TABLE_FOOTER - The footer associated with a table. A footer is typically
  	//    a line or lines of text below a table or embedded as the last row of a
  	//    table.
  	//
  	//    * CELL - A cell within a detected table. The cell is the parent of the
  	//    block that contains the text in the cell.
  	//
  	//    * MERGED_CELL - A cell in a table whose content spans more than one row
  	//    or column. The Relationships array for this cell contain data from individual
  	//    cells.
  	//
  	//    * SELECTION_ELEMENT - A selection element such as an option button (radio
  	//    button) or a check box that's detected on a document page. Use the value
  	//    of SelectionStatus to determine the status of the selection element.
  	//
  	//    * SIGNATURE - The location and confidence score of a signature detected
  	//    on a document page. Can be returned as part of a Key-Value pair or a detected
  	//    cell.
  	//
  	//    * QUERY - A question asked during the call of AnalyzeDocument. Contains
  	//    an alias and an ID that attaches it to its answer.
  	//
  	//    * QUERY_RESULT - A response to a question asked during the call of analyze
  	//    document. Comes with an alias and ID for ease of locating in a response.
  	//    Also contains location and confidence score.
  	//
  	// The following BlockTypes are only returned for Amazon Textract Layout.
  	//
  	//    * LAYOUT_TITLE - The main title of the document.
  	//
  	//    * LAYOUT_HEADER - Text located in the top margin of the document.
  	//
  	//    * LAYOUT_FOOTER - Text located in the bottom margin of the document.
  	//
  	//    * LAYOUT_SECTION_HEADER - The titles of sections within a document.
  	//
  	//    * LAYOUT_PAGE_NUMBER - The page number of the documents.
  	//
  	//    * LAYOUT_LIST - Any information grouped together in list form.
  	//
  	//    * LAYOUT_FIGURE - Indicates the location of an image in a document.
  	//
  	//    * LAYOUT_TABLE - Indicates the location of a table in the document.
  	//
  	//    * LAYOUT_KEY_VALUE - Indicates the location of form key-values in a document.
  	//
  	//    * LAYOUT_TEXT - Text that is present typically as a part of paragraphs
  	//    in documents.
  	BlockType *string `type:"string" enum:"BlockType"`
  
  	// The column in which a table cell appears. The first column position is 1.
  	// ColumnIndex isn't returned by DetectDocumentText and GetDocumentTextDetection.
  	ColumnIndex *int64 `type:"integer"`
  
  	// The number of columns that a table cell spans. ColumnSpan isn't returned
  	// by DetectDocumentText and GetDocumentTextDetection.
  	ColumnSpan *int64 `type:"integer"`
  
  	// The confidence score that Amazon Textract has in the accuracy of the recognized
  	// text and the accuracy of the geometry points around the recognized text.
  	Confidence *float64 `type:"float"`
  
  	// The type of entity.
  	//
  	// The following entity types can be returned by FORMS analysis:
  	//
  	//    * KEY - An identifier for a field on the document.
  	//
  	//    * VALUE - The field text.
  	//
  	// The following entity types can be returned by TABLES analysis:
  	//
  	//    * COLUMN_HEADER - Identifies a cell that is a header of a column.
  	//
  	//    * TABLE_TITLE - Identifies a cell that is a title within the table.
  	//
  	//    * TABLE_SECTION_TITLE - Identifies a cell that is a title of a section
  	//    within a table. A section title is a cell that typically spans an entire
  	//    row above a section.
  	//
  	//    * TABLE_FOOTER - Identifies a cell that is a footer of a table.
  	//
  	//    * TABLE_SUMMARY - Identifies a summary cell of a table. A summary cell
  	//    can be a row of a table or an additional, smaller table that contains
  	//    summary information for another table.
  	//
  	//    * STRUCTURED_TABLE - Identifies a table with column headers where the
  	//    content of each row corresponds to the headers.
  	//
  	//    * SEMI_STRUCTURED_TABLE - Identifies a non-structured table.
  	//
  	// EntityTypes isn't returned by DetectDocumentText and GetDocumentTextDetection.
  	EntityTypes []*string `type:"list" enum:"EntityType"`
  
  	// The location of the recognized text on the image. It includes an axis-aligned,
  	// coarse bounding box that surrounds the text, and a finer-grain polygon for
  	// more accurate spatial information.
  	Geometry *Geometry `type:"structure"`
  
  	// The identifier for the recognized text. The identifier is only unique for
  	// a single operation.
  	Id *string `type:"string"`
  
  	// The page on which a block was detected. Page is returned by synchronous and
  	// asynchronous operations. Page values greater than 1 are only returned for
  	// multipage documents that are in PDF or TIFF format. A scanned image (JPEG/PNG)
  	// provided to an asynchronous operation, even if it contains multiple document
  	// pages, is considered a single-page document. This means that for scanned
  	// images the value of Page is always 1.
  	Page *int64 `type:"integer"`
  
  	// Each query contains the question you want to ask in the Text and the alias
  	// you want to associate.
  	Query *Query `type:"structure"`
  
  	// A list of relationship objects that describe how blocks are related to each
  	// other. For example, a LINE block object contains a CHILD relationship type
  	// with the WORD blocks that make up the line of text. There aren't Relationship
  	// objects in the list for relationships that don't exist, such as when the
  	// current block has no child blocks.
  	Relationships []*Relationship `type:"list"`
  
  	// The row in which a table cell is located. The first row position is 1. RowIndex
  	// isn't returned by DetectDocumentText and GetDocumentTextDetection.
  	RowIndex *int64 `type:"integer"`
  
  	// The number of rows that a table cell spans. RowSpan isn't returned by DetectDocumentText
  	// and GetDocumentTextDetection.
  	RowSpan *int64 `type:"integer"`
  
  	// The selection status of a selection element, such as an option button or
  	// check box.
  	SelectionStatus *string `type:"string" enum:"SelectionStatus"`
  
  	// The word or line of text that's recognized by Amazon Textract.
  	Text *string `type:"string"`
  
  	// The kind of text that Amazon Textract has detected. Can check for handwritten
  	// text and printed text.
  	TextType *string `type:"string" enum:"TextType"`
  }

  // The bounding box around the detected page, text, key-value pair, table, table
// cell, or selection element on a document page. The left (x-coordinate) and
// top (y-coordinate) are coordinates that represent the top and left sides
// of the bounding box. Note that the upper-left corner of the image is the
// origin (0,0).
//
// The top and left values returned are ratios of the overall document page
// size. For example, if the input image is 700 x 200 pixels, and the top-left
// coordinate of the bounding box is 350 x 50 pixels, the API returns a left
// value of 0.5 (350/700) and a top value of 0.25 (50/200).
//
// The width and height values represent the dimensions of the bounding box
// as a ratio of the overall document page dimension. For example, if the document
// page size is 700 x 200 pixels, and the bounding box width is 70 pixels, the
// width returned is 0.1. 
  type BoundingBox   struct {
  	_ struct{} `type:"structure"`
  
  	// The height of the bounding box as a ratio of the overall document page height.
  	Height *float64 `type:"float"`
  
  	// The left coordinate of the bounding box as a ratio of overall document page
  	// width.
  	Left *float64 `type:"float"`
  
  	// The top coordinate of the bounding box as a ratio of overall document page
  	// height.
  	Top *float64 `type:"float"`
  
  	// The width of the bounding box as a ratio of the overall document page width.
  	Width *float64 `type:"float"`
  }

  // Updating or deleting a resource can cause an inconsistent state. 
  type ConflictException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  
  type CreateAdapterInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The name to be assigned to the adapter being created.
  	//
  	// AdapterName is a required field
  	AdapterName *string `min:"1" type:"string" required:"true"`
  
  	// Controls whether or not the adapter should automatically update.
  	AutoUpdate *string `type:"string" enum:"AutoUpdate"`
  
  	// Idempotent token is used to recognize the request. If the same token is used
  	// with multiple CreateAdapter requests, the same session is returned. This
  	// token is employed to avoid unintentionally creating the same session multiple
  	// times.
  	ClientRequestToken *string `min:"1" type:"string" idempotencyToken:"true"`
  
  	// The description to be assigned to the adapter being created.
  	Description *string `min:"1" type:"string"`
  
  	// The type of feature that the adapter is being trained on. Currrenly, supported
  	// feature types are: QUERIES
  	//
  	// FeatureTypes is a required field
  	FeatureTypes []*string `type:"list" required:"true" enum:"FeatureType"`
  
  	// A list of tags to be added to the adapter.
  	Tags map[string]*string `type:"map"`
  }

  
  type CreateAdapterOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing the unique ID for the adapter that has been created.
  	AdapterId *string `min:"12" type:"string"`
  }

  
  type CreateAdapterVersionInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter that will receive a new version.
  	//
  	// AdapterId is a required field
  	AdapterId *string `min:"12" type:"string" required:"true"`
  
  	// Idempotent token is used to recognize the request. If the same token is used
  	// with multiple CreateAdapterVersion requests, the same session is returned.
  	// This token is employed to avoid unintentionally creating the same session
  	// multiple times.
  	ClientRequestToken *string `min:"1" type:"string" idempotencyToken:"true"`
  
  	// Specifies a dataset used to train a new adapter version. Takes a ManifestS3Object
  	// as the value.
  	//
  	// DatasetConfig is a required field
  	DatasetConfig *AdapterVersionDatasetConfig `type:"structure" required:"true"`
  
  	// The identifier for your AWS Key Management Service key (AWS KMS key). Used
  	// to encrypt your documents.
  	KMSKeyId *string `min:"1" type:"string"`
  
  	// Sets whether or not your output will go to a user created bucket. Used to
  	// set the name of the bucket, and the prefix on the output file.
  	//
  	// OutputConfig is an optional parameter which lets you adjust where your output
  	// will be placed. By default, Amazon Textract will store the results internally
  	// and can only be accessed by the Get API operations. With OutputConfig enabled,
  	// you can set the name of the bucket the output will be sent to the file prefix
  	// of the results where you can download your results. Additionally, you can
  	// set the KMSKeyID parameter to a customer master key (CMK) to encrypt your
  	// output. Without this parameter set Amazon Textract will encrypt server-side
  	// using the AWS managed CMK for Amazon S3.
  	//
  	// Decryption of Customer Content is necessary for processing of the documents
  	// by Amazon Textract. If your account is opted out under an AI services opt
  	// out policy then all unencrypted Customer Content is immediately and permanently
  	// deleted after the Customer Content has been processed by the service. No
  	// copy of of the output is retained by Amazon Textract. For information about
  	// how to opt out, see Managing AI services opt-out policy. (https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_ai-opt-out.html)
  	//
  	// For more information on data privacy, see the Data Privacy FAQ (https://aws.amazon.com/compliance/data-privacy-faq/).
  	//
  	// OutputConfig is a required field
  	OutputConfig *OutputConfig `type:"structure" required:"true"`
  
  	// A set of tags (key-value pairs) that you want to attach to the adapter version.
  	Tags map[string]*string `type:"map"`
  }

  
  type CreateAdapterVersionOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing the unique ID for the adapter that has received a new
  	// version.
  	AdapterId *string `min:"12" type:"string"`
  
  	// A string describing the new version of the adapter.
  	AdapterVersion *string `min:"1" type:"string"`
  }

  
  type DeleteAdapterInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter to be deleted.
  	//
  	// AdapterId is a required field
  	AdapterId *string `min:"12" type:"string" required:"true"`
  }

  
  type DeleteAdapterOutput   struct {
  	_ struct{} `type:"structure"`
  }

  
  type DeleteAdapterVersionInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter version that will be deleted.
  	//
  	// AdapterId is a required field
  	AdapterId *string `min:"12" type:"string" required:"true"`
  
  	// Specifies the adapter version to be deleted.
  	//
  	// AdapterVersion is a required field
  	AdapterVersion *string `min:"1" type:"string" required:"true"`
  }

  
  type DeleteAdapterVersionOutput   struct {
  	_ struct{} `type:"structure"`
  }

  
  type DetectDocumentTextInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The input document as base64-encoded bytes or an Amazon S3 object. If you
  	// use the AWS CLI to call Amazon Textract operations, you can't pass image
  	// bytes. The document must be an image in JPEG or PNG format.
  	//
  	// If you're using an AWS SDK to call Amazon Textract, you might not need to
  	// base64-encode image bytes that are passed using the Bytes field.
  	//
  	// Document is a required field
  	Document *Document `type:"structure" required:"true"`
  }

  
  type DetectDocumentTextOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// An array of Block objects that contain the text that's detected in the document.
  	Blocks []*Block `type:"list"`
  
  	DetectDocumentTextModelVersion *string `type:"string"`
  
  	// Metadata about the document. It contains the number of pages that are detected
  	// in the document.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  }

  // A structure that holds information regarding a detected signature on a page. 
  type DetectedSignature   struct {
  	_ struct{} `type:"structure"`
  
  	// The page a detected signature was found on.
  	Page *int64 `type:"integer"`
  }

  // The input document, either as bytes or as an S3 object.
//
// You pass image bytes to an Amazon Textract API operation by using the Bytes
// property. For example, you would use the Bytes property to pass a document
// loaded from a local file system. Image bytes passed by using the Bytes property
// must be base64 encoded. Your code might not need to encode document file
// bytes if you're using an AWS SDK to call Amazon Textract API operations.
//
// You pass images stored in an S3 bucket to an Amazon Textract API operation
// by using the S3Object property. Documents stored in an S3 bucket don't need
// to be base64 encoded.
//
// The AWS Region for the S3 bucket that contains the S3 object must match the
// AWS Region that you use for Amazon Textract operations.
//
// If you use the AWS CLI to call Amazon Textract operations, passing image
// bytes using the Bytes property isn't supported. You must first upload the
// document to an Amazon S3 bucket, and then call the operation using the S3Object
// property.
//
// For Amazon Textract to process an S3 object, the user must have permission
// to access the S3 object. 
  type Document   struct {
  	_ struct{} `type:"structure"`
  
  	// A blob of base64-encoded document bytes. The maximum size of a document that's
  	// provided in a blob of bytes is 5 MB. The document bytes must be in PNG or
  	// JPEG format.
  	//
  	// If you're using an AWS SDK to call Amazon Textract, you might not need to
  	// base64-encode image bytes passed using the Bytes field.
  	// Bytes is automatically base64 encoded/decoded by the SDK.
  	Bytes []byte `min:"1" type:"blob"`
  
  	// Identifies an S3 object as the document source. The maximum size of a document
  	// that's stored in an S3 bucket is 5 MB.
  	S3Object *S3Object `type:"structure"`
  }

  // Summary information about documents grouped by the same document type. 
  type DocumentGroup   struct {
  	_ struct{} `type:"structure"`
  
  	// A list of the detected signatures found in a document group.
  	DetectedSignatures []*DetectedSignature `type:"list"`
  
  	// An array that contains information about the pages of a document, defined
  	// by logical boundary.
  	SplitDocuments []*SplitDocument `type:"list"`
  
  	// The type of document that Amazon Textract has detected. See Analyze Lending
  	// Response Objects (https://docs.aws.amazon.com/textract/latest/dg/lending-response-objects.html)
  	// for a list of all types returned by Textract.
  	Type *string `type:"string"`
  
  	// A list of any expected signatures not found in a document group.
  	UndetectedSignatures []*UndetectedSignature `type:"list"`
  }

  // The Amazon S3 bucket that contains the document to be processed. It's used
// by asynchronous operations.
//
// The input document can be an image file in JPEG or PNG format. It can also
// be a file in PDF format. 
  type DocumentLocation   struct {
  	_ struct{} `type:"structure"`
  
  	// The Amazon S3 bucket that contains the input document.
  	S3Object *S3Object `type:"structure"`
  }

  // Information about the input document. 
  type DocumentMetadata   struct {
  	_ struct{} `type:"structure"`
  
  	// The number of pages that are detected in the document.
  	Pages *int64 `type:"integer"`
  }

  // The document can't be processed because it's too large. The maximum document
// size for synchronous operations 10 MB. The maximum document size for asynchronous
// operations is 500 MB for PDF files. 
  type DocumentTooLargeException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // The evaluation metrics (F1 score, Precision, and Recall) for an adapter version. 
  type EvaluationMetric   struct {
  	_ struct{} `type:"structure"`
  
  	// The F1 score for an adapter version.
  	F1Score *float64 `type:"float"`
  
  	// The Precision score for an adapter version.
  	Precision *float64 `type:"float"`
  
  	// The Recall score for an adapter version.
  	Recall *float64 `type:"float"`
  }

  // Returns the kind of currency detected. 
  type ExpenseCurrency   struct {
  	_ struct{} `type:"structure"`
  
  	// Currency code for detected currency. the current supported codes are:
  	//
  	//    * USD
  	//
  	//    * EUR
  	//
  	//    * GBP
  	//
  	//    * CAD
  	//
  	//    * INR
  	//
  	//    * JPY
  	//
  	//    * CHF
  	//
  	//    * AUD
  	//
  	//    * CNY
  	//
  	//    * BZR
  	//
  	//    * SEK
  	//
  	//    * HKD
  	Code *string `type:"string"`
  
  	// Percentage confideence in the detected currency.
  	Confidence *float64 `type:"float"`
  }

  // An object used to store information about the Value or Label detected by
// Amazon Textract. 
  type ExpenseDetection   struct {
  	_ struct{} `type:"structure"`
  
  	// The confidence in detection, as a percentage
  	Confidence *float64 `type:"float"`
  
  	// Information about where the following items are located on a document page:
  	// detected page, text, key-value pairs, tables, table cells, and selection
  	// elements.
  	Geometry *Geometry `type:"structure"`
  
  	// The word or line of text recognized by Amazon Textract
  	Text *string `type:"string"`
  }

  // The structure holding all the information returned by AnalyzeExpense 
  type ExpenseDocument   struct {
  	_ struct{} `type:"structure"`
  
  	// This is a block object, the same as reported when DetectDocumentText is run
  	// on a document. It provides word level recognition of text.
  	Blocks []*Block `type:"list"`
  
  	// Denotes which invoice or receipt in the document the information is coming
  	// from. First document will be 1, the second 2, and so on.
  	ExpenseIndex *int64 `type:"integer"`
  
  	// Information detected on each table of a document, seperated into LineItems.
  	LineItemGroups []*LineItemGroup `type:"list"`
  
  	// Any information found outside of a table by Amazon Textract.
  	SummaryFields []*ExpenseField `type:"list"`
  }

  // Breakdown of detected information, seperated into the catagories Type, LabelDetection,
// and ValueDetection 
  type ExpenseField   struct {
  	_ struct{} `type:"structure"`
  
  	// Shows the kind of currency, both the code and confidence associated with
  	// any monatary value detected.
  	Currency *ExpenseCurrency `type:"structure"`
  
  	// Shows which group a response object belongs to, such as whether an address
  	// line belongs to the vendor's address or the recipent's address.
  	GroupProperties []*ExpenseGroupProperty `type:"list"`
  
  	// The explicitly stated label of a detected element.
  	LabelDetection *ExpenseDetection `type:"structure"`
  
  	// The page number the value was detected on.
  	PageNumber *int64 `type:"integer"`
  
  	// The implied label of a detected element. Present alongside LabelDetection
  	// for explicit elements.
  	Type *ExpenseType `type:"structure"`
  
  	// The value of a detected element. Present in explicit and implicit elements.
  	ValueDetection *ExpenseDetection `type:"structure"`
  }

  // Shows the group that a certain key belongs to. This helps differentiate between
// names and addresses for different organizations, that can be hard to determine
// via JSON response. 
  type ExpenseGroupProperty   struct {
  	_ struct{} `type:"structure"`
  
  	// Provides a group Id number, which will be the same for each in the group.
  	Id *string `type:"string"`
  
  	// Informs you on whether the expense group is a name or an address.
  	Types []*string `type:"list"`
  }

  // An object used to store information about the Type detected by Amazon Textract. 
  type ExpenseType   struct {
  	_ struct{} `type:"structure"`
  
  	// The confidence of accuracy, as a percentage.
  	Confidence *float64 `type:"float"`
  
  	// The word or line of text detected by Amazon Textract.
  	Text *string `type:"string"`
  }

  // Contains information extracted by an analysis operation after using StartLendingAnalysis. 
  type Extraction   struct {
  	_ struct{} `type:"structure"`
  
  	// The structure holding all the information returned by AnalyzeExpense
  	ExpenseDocument *ExpenseDocument `type:"structure"`
  
  	// The structure that lists each document processed in an AnalyzeID operation.
  	IdentityDocument *IdentityDocument `type:"structure"`
  
  	// Holds the structured data returned by AnalyzeDocument for lending documents.
  	LendingDocument *LendingDocument `type:"structure"`
  }

  // Information about where the following items are located on a document page:
// detected page, text, key-value pairs, tables, table cells, and selection
// elements. 
  type Geometry   struct {
  	_ struct{} `type:"structure"`
  
  	// An axis-aligned coarse representation of the location of the recognized item
  	// on the document page.
  	BoundingBox *BoundingBox `type:"structure"`
  
  	// Within the bounding box, a fine-grained polygon around the recognized item.
  	Polygon []*Point `type:"list"`
  }

  
  type GetAdapterInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter.
  	//
  	// AdapterId is a required field
  	AdapterId *string `min:"12" type:"string" required:"true"`
  }

  
  type GetAdapterOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string identifying the adapter that information has been retrieved for.
  	AdapterId *string `min:"12" type:"string"`
  
  	// The name of the requested adapter.
  	AdapterName *string `min:"1" type:"string"`
  
  	// Binary value indicating if the adapter is being automatically updated or
  	// not.
  	AutoUpdate *string `type:"string" enum:"AutoUpdate"`
  
  	// The date and time the requested adapter was created at.
  	CreationTime *time.Time `type:"timestamp"`
  
  	// The description for the requested adapter.
  	Description *string `min:"1" type:"string"`
  
  	// List of the targeted feature types for the requested adapter.
  	FeatureTypes []*string `type:"list" enum:"FeatureType"`
  
  	// A set of tags (key-value pairs) associated with the adapter that has been
  	// retrieved.
  	Tags map[string]*string `type:"map"`
  }

  
  type GetAdapterVersionInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string specifying a unique ID for the adapter version you want to retrieve
  	// information for.
  	//
  	// AdapterId is a required field
  	AdapterId *string `min:"12" type:"string" required:"true"`
  
  	// A string specifying the adapter version you want to retrieve information
  	// for.
  	//
  	// AdapterVersion is a required field
  	AdapterVersion *string `min:"1" type:"string" required:"true"`
  }

  
  type GetAdapterVersionOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter version being retrieved.
  	AdapterId *string `min:"12" type:"string"`
  
  	// A string containing the adapter version that has been retrieved.
  	AdapterVersion *string `min:"1" type:"string"`
  
  	// The time that the adapter version was created.
  	CreationTime *time.Time `type:"timestamp"`
  
  	// Specifies a dataset used to train a new adapter version. Takes a ManifestS3Objec
  	// as the value.
  	DatasetConfig *AdapterVersionDatasetConfig `type:"structure"`
  
  	// The evaluation metrics (F1 score, Precision, and Recall) for the requested
  	// version, grouped by baseline metrics and adapter version.
  	EvaluationMetrics []*AdapterVersionEvaluationMetric `type:"list"`
  
  	// List of the targeted feature types for the requested adapter version.
  	FeatureTypes []*string `type:"list" enum:"FeatureType"`
  
  	// The identifier for your AWS Key Management Service key (AWS KMS key). Used
  	// to encrypt your documents.
  	KMSKeyId *string `min:"1" type:"string"`
  
  	// Sets whether or not your output will go to a user created bucket. Used to
  	// set the name of the bucket, and the prefix on the output file.
  	//
  	// OutputConfig is an optional parameter which lets you adjust where your output
  	// will be placed. By default, Amazon Textract will store the results internally
  	// and can only be accessed by the Get API operations. With OutputConfig enabled,
  	// you can set the name of the bucket the output will be sent to the file prefix
  	// of the results where you can download your results. Additionally, you can
  	// set the KMSKeyID parameter to a customer master key (CMK) to encrypt your
  	// output. Without this parameter set Amazon Textract will encrypt server-side
  	// using the AWS managed CMK for Amazon S3.
  	//
  	// Decryption of Customer Content is necessary for processing of the documents
  	// by Amazon Textract. If your account is opted out under an AI services opt
  	// out policy then all unencrypted Customer Content is immediately and permanently
  	// deleted after the Customer Content has been processed by the service. No
  	// copy of of the output is retained by Amazon Textract. For information about
  	// how to opt out, see Managing AI services opt-out policy. (https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_ai-opt-out.html)
  	//
  	// For more information on data privacy, see the Data Privacy FAQ (https://aws.amazon.com/compliance/data-privacy-faq/).
  	OutputConfig *OutputConfig `type:"structure"`
  
  	// The status of the adapter version that has been requested.
  	Status *string `type:"string" enum:"AdapterVersionStatus"`
  
  	// A message that describes the status of the requested adapter version.
  	StatusMessage *string `min:"1" type:"string"`
  
  	// A set of tags (key-value pairs) that are associated with the adapter version.
  	Tags map[string]*string `type:"map"`
  }

  
  type GetDocumentAnalysisInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the text-detection job. The JobId is returned from
  	// StartDocumentAnalysis. A JobId value is only valid for 7 days.
  	//
  	// JobId is a required field
  	JobId *string `min:"1" type:"string" required:"true"`
  
  	// The maximum number of results to return per paginated call. The largest value
  	// that you can specify is 1,000. If you specify a value greater than 1,000,
  	// a maximum of 1,000 results is returned. The default value is 1,000.
  	MaxResults *int64 `min:"1" type:"integer"`
  
  	// If the previous response was incomplete (because there are more blocks to
  	// retrieve), Amazon Textract returns a pagination token in the response. You
  	// can use this pagination token to retrieve the next set of blocks.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type GetDocumentAnalysisOutput   struct {
  	_ struct{} `type:"structure"`
  
  	AnalyzeDocumentModelVersion *string `type:"string"`
  
  	// The results of the text-analysis operation.
  	Blocks []*Block `type:"list"`
  
  	// Information about a document that Amazon Textract processed. DocumentMetadata
  	// is returned in every page of paginated responses from an Amazon Textract
  	// video operation.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// The current status of the text detection job.
  	JobStatus *string `type:"string" enum:"JobStatus"`
  
  	// If the response is truncated, Amazon Textract returns this token. You can
  	// use this token in the subsequent request to retrieve the next set of text
  	// detection results.
  	NextToken *string `min:"1" type:"string"`
  
  	// Returns if the detection job could not be completed. Contains explanation
  	// for what error occured.
  	StatusMessage *string `type:"string"`
  
  	// A list of warnings that occurred during the document-analysis operation.
  	Warnings []*Warning `type:"list"`
  }

  
  type GetDocumentTextDetectionInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the text detection job. The JobId is returned from
  	// StartDocumentTextDetection. A JobId value is only valid for 7 days.
  	//
  	// JobId is a required field
  	JobId *string `min:"1" type:"string" required:"true"`
  
  	// The maximum number of results to return per paginated call. The largest value
  	// you can specify is 1,000. If you specify a value greater than 1,000, a maximum
  	// of 1,000 results is returned. The default value is 1,000.
  	MaxResults *int64 `min:"1" type:"integer"`
  
  	// If the previous response was incomplete (because there are more blocks to
  	// retrieve), Amazon Textract returns a pagination token in the response. You
  	// can use this pagination token to retrieve the next set of blocks.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type GetDocumentTextDetectionOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The results of the text-detection operation.
  	Blocks []*Block `type:"list"`
  
  	DetectDocumentTextModelVersion *string `type:"string"`
  
  	// Information about a document that Amazon Textract processed. DocumentMetadata
  	// is returned in every page of paginated responses from an Amazon Textract
  	// video operation.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// The current status of the text detection job.
  	JobStatus *string `type:"string" enum:"JobStatus"`
  
  	// If the response is truncated, Amazon Textract returns this token. You can
  	// use this token in the subsequent request to retrieve the next set of text-detection
  	// results.
  	NextToken *string `min:"1" type:"string"`
  
  	// Returns if the detection job could not be completed. Contains explanation
  	// for what error occured.
  	StatusMessage *string `type:"string"`
  
  	// A list of warnings that occurred during the text-detection operation for
  	// the document.
  	Warnings []*Warning `type:"list"`
  }

  
  type GetExpenseAnalysisInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the text detection job. The JobId is returned from
  	// StartExpenseAnalysis. A JobId value is only valid for 7 days.
  	//
  	// JobId is a required field
  	JobId *string `min:"1" type:"string" required:"true"`
  
  	// The maximum number of results to return per paginated call. The largest value
  	// you can specify is 20. If you specify a value greater than 20, a maximum
  	// of 20 results is returned. The default value is 20.
  	MaxResults *int64 `min:"1" type:"integer"`
  
  	// If the previous response was incomplete (because there are more blocks to
  	// retrieve), Amazon Textract returns a pagination token in the response. You
  	// can use this pagination token to retrieve the next set of blocks.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type GetExpenseAnalysisOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The current model version of AnalyzeExpense.
  	AnalyzeExpenseModelVersion *string `type:"string"`
  
  	// Information about a document that Amazon Textract processed. DocumentMetadata
  	// is returned in every page of paginated responses from an Amazon Textract
  	// operation.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// The expenses detected by Amazon Textract.
  	ExpenseDocuments []*ExpenseDocument `type:"list"`
  
  	// The current status of the text detection job.
  	JobStatus *string `type:"string" enum:"JobStatus"`
  
  	// If the response is truncated, Amazon Textract returns this token. You can
  	// use this token in the subsequent request to retrieve the next set of text-detection
  	// results.
  	NextToken *string `min:"1" type:"string"`
  
  	// Returns if the detection job could not be completed. Contains explanation
  	// for what error occured.
  	StatusMessage *string `type:"string"`
  
  	// A list of warnings that occurred during the text-detection operation for
  	// the document.
  	Warnings []*Warning `type:"list"`
  }

  
  type GetLendingAnalysisInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the lending or text-detection job. The JobId is returned
  	// from StartLendingAnalysis. A JobId value is only valid for 7 days.
  	//
  	// JobId is a required field
  	JobId *string `min:"1" type:"string" required:"true"`
  
  	// The maximum number of results to return per paginated call. The largest value
  	// that you can specify is 30. If you specify a value greater than 30, a maximum
  	// of 30 results is returned. The default value is 30.
  	MaxResults *int64 `min:"1" type:"integer"`
  
  	// If the previous response was incomplete, Amazon Textract returns a pagination
  	// token in the response. You can use this pagination token to retrieve the
  	// next set of lending results.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type GetLendingAnalysisOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The current model version of the Analyze Lending API.
  	AnalyzeLendingModelVersion *string `type:"string"`
  
  	// Information about the input document.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// The current status of the lending analysis job.
  	JobStatus *string `type:"string" enum:"JobStatus"`
  
  	// If the response is truncated, Amazon Textract returns this token. You can
  	// use this token in the subsequent request to retrieve the next set of lending
  	// results.
  	NextToken *string `min:"1" type:"string"`
  
  	// Holds the information returned by one of AmazonTextract's document analysis
  	// operations for the pinstripe.
  	Results []*LendingResult `type:"list"`
  
  	// Returns if the lending analysis job could not be completed. Contains explanation
  	// for what error occurred.
  	StatusMessage *string `type:"string"`
  
  	// A list of warnings that occurred during the lending analysis operation.
  	Warnings []*Warning `type:"list"`
  }

  
  type GetLendingAnalysisSummaryInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the lending or text-detection job. The JobId is returned
  	// from StartLendingAnalysis. A JobId value is only valid for 7 days.
  	//
  	// JobId is a required field
  	JobId *string `min:"1" type:"string" required:"true"`
  }

  
  type GetLendingAnalysisSummaryOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The current model version of the Analyze Lending API.
  	AnalyzeLendingModelVersion *string `type:"string"`
  
  	// Information about the input document.
  	DocumentMetadata *DocumentMetadata `type:"structure"`
  
  	// The current status of the lending analysis job.
  	JobStatus *string `type:"string" enum:"JobStatus"`
  
  	// Returns if the lending analysis could not be completed. Contains explanation
  	// for what error occurred.
  	StatusMessage *string `type:"string"`
  
  	// Contains summary information for documents grouped by type.
  	Summary *LendingSummary `type:"structure"`
  
  	// A list of warnings that occurred during the lending analysis operation.
  	Warnings []*Warning `type:"list"`
  }

  // Shows the results of the human in the loop evaluation. If there is no HumanLoopArn,
// the input did not trigger human review. 
  type HumanLoopActivationOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// Shows the result of condition evaluations, including those conditions which
  	// activated a human review.
  	HumanLoopActivationConditionsEvaluationResults aws.JSONValue `type:"jsonvalue"`
  
  	// Shows if and why human review was needed.
  	HumanLoopActivationReasons []*string `min:"1" type:"list"`
  
  	// The Amazon Resource Name (ARN) of the HumanLoop created.
  	HumanLoopArn *string `type:"string"`
  }

  // Sets up the human review workflow the document will be sent to if one of
// the conditions is met. You can also set certain attributes of the image before
// review. 
  type HumanLoopConfig   struct {
  	_ struct{} `type:"structure"`
  
  	// Sets attributes of the input data.
  	DataAttributes *HumanLoopDataAttributes `type:"structure"`
  
  	// The Amazon Resource Name (ARN) of the flow definition.
  	//
  	// FlowDefinitionArn is a required field
  	FlowDefinitionArn *string `type:"string" required:"true"`
  
  	// The name of the human workflow used for this image. This should be kept unique
  	// within a region.
  	//
  	// HumanLoopName is a required field
  	HumanLoopName *string `min:"1" type:"string" required:"true"`
  }

  // Allows you to set attributes of the image. Currently, you can declare an
// image as free of personally identifiable information and adult content. 
  type HumanLoopDataAttributes   struct {
  	_ struct{} `type:"structure"`
  
  	// Sets whether the input image is free of personally identifiable information
  	// or adult content.
  	ContentClassifiers []*string `type:"list" enum:"ContentClassifier"`
  }

  // Indicates you have exceeded the maximum number of active human in the loop
// workflows available 
  type HumanLoopQuotaExceededException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  
  	// The quota code.
  	QuotaCode *string `type:"string"`
  
  	// The resource type.
  	ResourceType *string `type:"string"`
  
  	// The service code.
  	ServiceCode *string `type:"string"`
  }

  // A ClientRequestToken input parameter was reused with an operation, but at
// least one of the other input parameters is different from the previous call
// to the operation. 
  type IdempotentParameterMismatchException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // The structure that lists each document processed in an AnalyzeID operation. 
  type IdentityDocument   struct {
  	_ struct{} `type:"structure"`
  
  	// Individual word recognition, as returned by document detection.
  	Blocks []*Block `type:"list"`
  
  	// Denotes the placement of a document in the IdentityDocument list. The first
  	// document is marked 1, the second 2 and so on.
  	DocumentIndex *int64 `type:"integer"`
  
  	// The structure used to record information extracted from identity documents.
  	// Contains both normalized field and value of the extracted text.
  	IdentityDocumentFields []*IdentityDocumentField `type:"list"`
  }

  // Structure containing both the normalized type of the extracted information
// and the text associated with it. These are extracted as Type and Value respectively. 
  type IdentityDocumentField   struct {
  	_ struct{} `type:"structure"`
  
  	// Used to contain the information detected by an AnalyzeID operation.
  	Type *AnalyzeIDDetections `type:"structure"`
  
  	// Used to contain the information detected by an AnalyzeID operation.
  	ValueDetection *AnalyzeIDDetections `type:"structure"`
  }

  // Amazon Textract experienced a service issue. Try your call again. 
  type InternalServerError   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // An invalid job identifier was passed to an asynchronous analysis operation. 
  type InvalidJobIdException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // Indicates you do not have decrypt permissions with the KMS key entered, or
// the KMS key was entered incorrectly. 
  type InvalidKMSKeyException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // An input parameter violated a constraint. For example, in synchronous operations,
// an InvalidParameterException exception occurs when neither of the S3Object
// or Bytes values are supplied in the Document request parameter. Validate
// your parameter before calling the API operation again. 
  type InvalidParameterException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // Amazon Textract is unable to access the S3 object that's specified in the
// request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
// For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html) 
  type InvalidS3ObjectException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // The results extracted for a lending document. 
  type LendingDetection   struct {
  	_ struct{} `type:"structure"`
  
  	// The confidence level for the text of a detected value in a lending document.
  	Confidence *float64 `type:"float"`
  
  	// Information about where the following items are located on a document page:
  	// detected page, text, key-value pairs, tables, table cells, and selection
  	// elements.
  	Geometry *Geometry `type:"structure"`
  
  	// The selection status of a selection element, such as an option button or
  	// check box.
  	SelectionStatus *string `type:"string" enum:"SelectionStatus"`
  
  	// The text extracted for a detected value in a lending document.
  	Text *string `type:"string"`
  }

  // Holds the structured data returned by AnalyzeDocument for lending documents. 
  type LendingDocument   struct {
  	_ struct{} `type:"structure"`
  
  	// An array of LendingField objects.
  	LendingFields []*LendingField `type:"list"`
  
  	// A list of signatures detected in a lending document.
  	SignatureDetections []*SignatureDetection `type:"list"`
  }

  // Holds the normalized key-value pairs returned by AnalyzeDocument, including
// the document type, detected text, and geometry. 
  type LendingField   struct {
  	_ struct{} `type:"structure"`
  
  	// The results extracted for a lending document.
  	KeyDetection *LendingDetection `type:"structure"`
  
  	// The type of the lending document.
  	Type *string `type:"string"`
  
  	// An array of LendingDetection objects.
  	ValueDetections []*LendingDetection `type:"list"`
  }

  // Contains the detections for each page analyzed through the Analyze Lending
// API. 
  type LendingResult   struct {
  	_ struct{} `type:"structure"`
  
  	// An array of Extraction to hold structured data. e.g. normalized key value
  	// pairs instead of raw OCR detections .
  	Extractions []*Extraction `type:"list"`
  
  	// The page number for a page, with regard to whole submission.
  	Page *int64 `type:"integer"`
  
  	// The classifier result for a given page.
  	PageClassification *PageClassification `type:"structure"`
  }

  // Contains information regarding DocumentGroups and UndetectedDocumentTypes. 
  type LendingSummary   struct {
  	_ struct{} `type:"structure"`
  
  	// Contains an array of all DocumentGroup objects.
  	DocumentGroups []*DocumentGroup `type:"list"`
  
  	// UndetectedDocumentTypes.
  	UndetectedDocumentTypes []*string `type:"list"`
  }

  // An Amazon Textract service limit was exceeded. For example, if you start
// too many asynchronous jobs concurrently, calls to start operations (StartDocumentTextDetection,
// for example) raise a LimitExceededException exception (HTTP status code:
// 400) until the number of concurrently running jobs is below the Amazon Textract
// service limit. 
  type LimitExceededException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // A structure that holds information about the different lines found in a document's
// tables. 
  type LineItemFields   struct {
  	_ struct{} `type:"structure"`
  
  	// ExpenseFields used to show information from detected lines on a table.
  	LineItemExpenseFields []*ExpenseField `type:"list"`
  }

  // A grouping of tables which contain LineItems, with each table identified
// by the table's LineItemGroupIndex. 
  type LineItemGroup   struct {
  	_ struct{} `type:"structure"`
  
  	// The number used to identify a specific table in a document. The first table
  	// encountered will have a LineItemGroupIndex of 1, the second 2, etc.
  	LineItemGroupIndex *int64 `type:"integer"`
  
  	// The breakdown of information on a particular line of a table.
  	LineItems []*LineItemFields `type:"list"`
  }

  
  type ListAdapterVersionsInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter to match for when listing
  	// adapter versions.
  	AdapterId *string `min:"12" type:"string"`
  
  	// Specifies the lower bound for the ListAdapterVersions operation. Ensures
  	// ListAdapterVersions returns only adapter versions created after the specified
  	// creation time.
  	AfterCreationTime *time.Time `type:"timestamp"`
  
  	// Specifies the upper bound for the ListAdapterVersions operation. Ensures
  	// ListAdapterVersions returns only adapter versions created after the specified
  	// creation time.
  	BeforeCreationTime *time.Time `type:"timestamp"`
  
  	// The maximum number of results to return when listing adapter versions.
  	MaxResults *int64 `min:"1" type:"integer"`
  
  	// Identifies the next page of results to return when listing adapter versions.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type ListAdapterVersionsOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// Adapter versions that match the filtering criteria specified when calling
  	// ListAdapters.
  	AdapterVersions []*AdapterVersionOverview `type:"list"`
  
  	// Identifies the next page of results to return when listing adapter versions.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type ListAdaptersInput   struct {
  	_ struct{} `type:"structure"`
  
  	// Specifies the lower bound for the ListAdapters operation. Ensures ListAdapters
  	// returns only adapters created after the specified creation time.
  	AfterCreationTime *time.Time `type:"timestamp"`
  
  	// Specifies the upper bound for the ListAdapters operation. Ensures ListAdapters
  	// returns only adapters created before the specified creation time.
  	BeforeCreationTime *time.Time `type:"timestamp"`
  
  	// The maximum number of results to return when listing adapters.
  	MaxResults *int64 `min:"1" type:"integer"`
  
  	// Identifies the next page of results to return when listing adapters.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type ListAdaptersOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A list of adapters that matches the filtering criteria specified when calling
  	// ListAdapters.
  	Adapters []*AdapterOverview `type:"list"`
  
  	// Identifies the next page of results to return when listing adapters.
  	NextToken *string `min:"1" type:"string"`
  }

  
  type ListTagsForResourceInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The Amazon Resource Name (ARN) that specifies the resource to list tags for.
  	//
  	// ResourceARN is a required field
  	ResourceARN *string `min:"1" type:"string" required:"true"`
  }

  
  type ListTagsForResourceOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A set of tags (key-value pairs) that are part of the requested resource.
  	Tags map[string]*string `type:"map"`
  }

  // Contains information relating to dates in a document, including the type
// of value, and the value. 
  type NormalizedValue   struct {
  	_ struct{} `type:"structure"`
  
  	// The value of the date, written as Year-Month-DayTHour:Minute:Second.
  	Value *string `type:"string"`
  
  	// The normalized type of the value detected. In this case, DATE.
  	ValueType *string `type:"string" enum:"ValueType"`
  }

  // The Amazon Simple Notification Service (Amazon SNS) topic to which Amazon
// Textract publishes the completion status of an asynchronous document operation. 
  type NotificationChannel   struct {
  	_ struct{} `type:"structure"`
  
  	// The Amazon Resource Name (ARN) of an IAM role that gives Amazon Textract
  	// publishing permissions to the Amazon SNS topic.
  	//
  	// RoleArn is a required field
  	RoleArn *string `min:"20" type:"string" required:"true"`
  
  	// The Amazon SNS topic that Amazon Textract posts the completion status to.
  	//
  	// SNSTopicArn is a required field
  	SNSTopicArn *string `min:"20" type:"string" required:"true"`
  }

  // Sets whether or not your output will go to a user created bucket. Used to
// set the name of the bucket, and the prefix on the output file.
//
// OutputConfig is an optional parameter which lets you adjust where your output
// will be placed. By default, Amazon Textract will store the results internally
// and can only be accessed by the Get API operations. With OutputConfig enabled,
// you can set the name of the bucket the output will be sent to the file prefix
// of the results where you can download your results. Additionally, you can
// set the KMSKeyID parameter to a customer master key (CMK) to encrypt your
// output. Without this parameter set Amazon Textract will encrypt server-side
// using the AWS managed CMK for Amazon S3.
//
// Decryption of Customer Content is necessary for processing of the documents
// by Amazon Textract. If your account is opted out under an AI services opt
// out policy then all unencrypted Customer Content is immediately and permanently
// deleted after the Customer Content has been processed by the service. No
// copy of of the output is retained by Amazon Textract. For information about
// how to opt out, see Managing AI services opt-out policy. (https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_ai-opt-out.html)
//
// For more information on data privacy, see the Data Privacy FAQ (https://aws.amazon.com/compliance/data-privacy-faq/). 
  type OutputConfig   struct {
  	_ struct{} `type:"structure"`
  
  	// The name of the bucket your output will go to.
  	//
  	// S3Bucket is a required field
  	S3Bucket *string `min:"3" type:"string" required:"true"`
  
  	// The prefix of the object key that the output will be saved to. When not enabled,
  	// the prefix will be “textract_output".
  	S3Prefix *string `min:"1" type:"string"`
  }

  // The class assigned to a Page object detected in an input document. Contains
// information regarding the predicted type/class of a document's page and the
// page number that the Page object was detected on. 
  type PageClassification   struct {
  	_ struct{} `type:"structure"`
  
  	// The page number the value was detected on, relative to Amazon Textract's
  	// starting position.
  	//
  	// PageNumber is a required field
  	PageNumber []*Prediction `type:"list" required:"true"`
  
  	// The class, or document type, assigned to a detected Page object. The class,
  	// or document type, assigned to a detected Page object.
  	//
  	// PageType is a required field
  	PageType []*Prediction `type:"list" required:"true"`
  }

  // The X and Y coordinates of a point on a document page. The X and Y values
// that are returned are ratios of the overall document page size. For example,
// if the input document is 700 x 200 and the operation returns X=0.5 and Y=0.25,
// then the point is at the (350,50) pixel coordinate on the document page.
//
// An array of Point objects, Polygon, is returned by DetectDocumentText. Polygon
// represents a fine-grained polygon around detected text. For more information,
// see Geometry in the Amazon Textract Developer Guide. 
  type Point   struct {
  	_ struct{} `type:"structure"`
  
  	// The value of the X coordinate for a point on a Polygon.
  	X *float64 `type:"float"`
  
  	// The value of the Y coordinate for a point on a Polygon.
  	Y *float64 `type:"float"`
  }

  // Contains information regarding predicted values returned by Amazon Textract
// operations, including the predicted value and the confidence in the predicted
// value. 
  type Prediction   struct {
  	_ struct{} `type:"structure"`
  
  	// Amazon Textract's confidence in its predicted value.
  	Confidence *float64 `type:"float"`
  
  	// The predicted value of a detected object.
  	Value *string `type:"string"`
  }

  // The number of requests exceeded your throughput limit. If you want to increase
// this limit, contact Amazon Textract. 
  type ProvisionedThroughputExceededException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  
  type QueriesConfig   struct {
  	_ struct{} `type:"structure"`
  
  	// Queries is a required field
  	Queries []*Query `min:"1" type:"list" required:"true"`
  }

  // Each query contains the question you want to ask in the Text and the alias
// you want to associate. 
  type Query   struct {
  	_ struct{} `type:"structure"`
  
  	// Alias attached to the query, for ease of location.
  	Alias *string `min:"1" type:"string"`
  
  	// Pages is a parameter that the user inputs to specify which pages to apply
  	// a query to. The following is a list of rules for using this parameter.
  	//
  	//    * If a page is not specified, it is set to ["1"] by default.
  	//
  	//    * The following characters are allowed in the parameter's string: 0 1
  	//    2 3 4 5 6 7 8 9 - *. No whitespace is allowed.
  	//
  	//    * When using * to indicate all pages, it must be the only element in the
  	//    list.
  	//
  	//    * You can use page intervals, such as [“1-3”, “1-1”, “4-*”].
  	//    Where * indicates last page of document.
  	//
  	//    * Specified pages must be greater than 0 and less than or equal to the
  	//    number of pages in the document.
  	Pages []*string `min:"1" type:"list"`
  
  	// Question that Amazon Textract will apply to the document. An example would
  	// be "What is the customer's SSN?"
  	//
  	// Text is a required field
  	Text *string `min:"1" type:"string" required:"true"`
  }

  // Information about how blocks are related to each other. A Block object contains
// 0 or more Relation objects in a list, Relationships. For more information,
// see Block.
//
// The Type element provides the type of the relationship for all blocks in
// the IDs array. 
  type Relationship   struct {
  	_ struct{} `type:"structure"`
  
  	// An array of IDs for related blocks. You can get the type of the relationship
  	// from the Type element.
  	Ids []*string `type:"list"`
  
  	// The type of relationship between the blocks in the IDs array and the current
  	// block. The following list describes the relationship types that can be returned.
  	//
  	//    * VALUE - A list that contains the ID of the VALUE block that's associated
  	//    with the KEY of a key-value pair.
  	//
  	//    * CHILD - A list of IDs that identify blocks found within the current
  	//    block object. For example, WORD blocks have a CHILD relationship to the
  	//    LINE block type.
  	//
  	//    * MERGED_CELL - A list of IDs that identify each of the MERGED_CELL block
  	//    types in a table.
  	//
  	//    * ANSWER - A list that contains the ID of the QUERY_RESULT block that’s
  	//    associated with the corresponding QUERY block.
  	//
  	//    * TABLE - A list of IDs that identify associated TABLE block types.
  	//
  	//    * TABLE_TITLE - A list that contains the ID for the TABLE_TITLE block
  	//    type in a table.
  	//
  	//    * TABLE_FOOTER - A list of IDs that identify the TABLE_FOOTER block types
  	//    in a table.
  	Type *string `type:"string" enum:"RelationshipType"`
  }

  // Returned when an operation tried to access a nonexistent resource. 
  type ResourceNotFoundException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // The S3 bucket name and file name that identifies the document.
//
// The AWS Region for the S3 bucket that contains the document must match the
// Region that you use for Amazon Textract operations.
//
// For Amazon Textract to process a file in an S3 bucket, the user must have
// permission to access the S3 bucket and file. 
  type S3Object   struct {
  	_ struct{} `type:"structure"`
  
  	// The name of the S3 bucket. Note that the # character is not valid in the
  	// file name.
  	Bucket *string `min:"3" type:"string"`
  
  	// The file name of the input document. Synchronous operations can use image
  	// files that are in JPEG or PNG format. Asynchronous operations also support
  	// PDF and TIFF format files.
  	Name *string `min:"1" type:"string"`
  
  	// If the bucket has versioning enabled, you can specify the object version.
  	Version *string `min:"1" type:"string"`
  }

  // Returned when a request cannot be completed as it would exceed a maximum
// service quota. 
  type ServiceQuotaExceededException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // Information regarding a detected signature on a page. 
  type SignatureDetection   struct {
  	_ struct{} `type:"structure"`
  
  	// The confidence, from 0 to 100, in the predicted values for a detected signature.
  	Confidence *float64 `type:"float"`
  
  	// Information about where the following items are located on a document page:
  	// detected page, text, key-value pairs, tables, table cells, and selection
  	// elements.
  	Geometry *Geometry `type:"structure"`
  }

  // Contains information about the pages of a document, defined by logical boundary. 
  type SplitDocument   struct {
  	_ struct{} `type:"structure"`
  
  	// The index for a given document in a DocumentGroup of a specific Type.
  	Index *int64 `type:"integer"`
  
  	// An array of page numbers for a for a given document, ordered by logical boundary.
  	Pages []*int64 `type:"list"`
  }

  
  type StartDocumentAnalysisInput   struct {
  	_ struct{} `type:"structure"`
  
  	// Specifies the adapter to be used when analyzing a document.
  	AdaptersConfig *AdaptersConfig `type:"structure"`
  
  	// The idempotent token that you use to identify the start request. If you use
  	// the same token with multiple StartDocumentAnalysis requests, the same JobId
  	// is returned. Use ClientRequestToken to prevent the same job from being accidentally
  	// started more than once. For more information, see Calling Amazon Textract
  	// Asynchronous Operations (https://docs.aws.amazon.com/textract/latest/dg/api-async.html).
  	ClientRequestToken *string `min:"1" type:"string"`
  
  	// The location of the document to be processed.
  	//
  	// DocumentLocation is a required field
  	DocumentLocation *DocumentLocation `type:"structure" required:"true"`
  
  	// A list of the types of analysis to perform. Add TABLES to the list to return
  	// information about the tables that are detected in the input document. Add
  	// FORMS to return detected form data. To perform both types of analysis, add
  	// TABLES and FORMS to FeatureTypes. All lines and words detected in the document
  	// are included in the response (including text that isn't related to the value
  	// of FeatureTypes).
  	//
  	// FeatureTypes is a required field
  	FeatureTypes []*string `type:"list" required:"true" enum:"FeatureType"`
  
  	// An identifier that you specify that's included in the completion notification
  	// published to the Amazon SNS topic. For example, you can use JobTag to identify
  	// the type of document that the completion notification corresponds to (such
  	// as a tax form or a receipt).
  	JobTag *string `min:"1" type:"string"`
  
  	// The KMS key used to encrypt the inference results. This can be in either
  	// Key ID or Key Alias format. When a KMS key is provided, the KMS key will
  	// be used for server-side encryption of the objects in the customer bucket.
  	// When this parameter is not enabled, the result will be encrypted server side,using
  	// SSE-S3.
  	KMSKeyId *string `min:"1" type:"string"`
  
  	// The Amazon SNS topic ARN that you want Amazon Textract to publish the completion
  	// status of the operation to.
  	NotificationChannel *NotificationChannel `type:"structure"`
  
  	// Sets if the output will go to a customer defined bucket. By default, Amazon
  	// Textract will save the results internally to be accessed by the GetDocumentAnalysis
  	// operation.
  	OutputConfig *OutputConfig `type:"structure"`
  
  	QueriesConfig *QueriesConfig `type:"structure"`
  }

  
  type StartDocumentAnalysisOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The identifier for the document text detection job. Use JobId to identify
  	// the job in a subsequent call to GetDocumentAnalysis. A JobId value is only
  	// valid for 7 days.
  	JobId *string `min:"1" type:"string"`
  }

  
  type StartDocumentTextDetectionInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The idempotent token that's used to identify the start request. If you use
  	// the same token with multiple StartDocumentTextDetection requests, the same
  	// JobId is returned. Use ClientRequestToken to prevent the same job from being
  	// accidentally started more than once. For more information, see Calling Amazon
  	// Textract Asynchronous Operations (https://docs.aws.amazon.com/textract/latest/dg/api-async.html).
  	ClientRequestToken *string `min:"1" type:"string"`
  
  	// The location of the document to be processed.
  	//
  	// DocumentLocation is a required field
  	DocumentLocation *DocumentLocation `type:"structure" required:"true"`
  
  	// An identifier that you specify that's included in the completion notification
  	// published to the Amazon SNS topic. For example, you can use JobTag to identify
  	// the type of document that the completion notification corresponds to (such
  	// as a tax form or a receipt).
  	JobTag *string `min:"1" type:"string"`
  
  	// The KMS key used to encrypt the inference results. This can be in either
  	// Key ID or Key Alias format. When a KMS key is provided, the KMS key will
  	// be used for server-side encryption of the objects in the customer bucket.
  	// When this parameter is not enabled, the result will be encrypted server side,using
  	// SSE-S3.
  	KMSKeyId *string `min:"1" type:"string"`
  
  	// The Amazon SNS topic ARN that you want Amazon Textract to publish the completion
  	// status of the operation to.
  	NotificationChannel *NotificationChannel `type:"structure"`
  
  	// Sets if the output will go to a customer defined bucket. By default Amazon
  	// Textract will save the results internally to be accessed with the GetDocumentTextDetection
  	// operation.
  	OutputConfig *OutputConfig `type:"structure"`
  }

  
  type StartDocumentTextDetectionOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// The identifier of the text detection job for the document. Use JobId to identify
  	// the job in a subsequent call to GetDocumentTextDetection. A JobId value is
  	// only valid for 7 days.
  	JobId *string `min:"1" type:"string"`
  }

  
  type StartExpenseAnalysisInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The idempotent token that's used to identify the start request. If you use
  	// the same token with multiple StartDocumentTextDetection requests, the same
  	// JobId is returned. Use ClientRequestToken to prevent the same job from being
  	// accidentally started more than once. For more information, see Calling Amazon
  	// Textract Asynchronous Operations (https://docs.aws.amazon.com/textract/latest/dg/api-async.html)
  	ClientRequestToken *string `min:"1" type:"string"`
  
  	// The location of the document to be processed.
  	//
  	// DocumentLocation is a required field
  	DocumentLocation *DocumentLocation `type:"structure" required:"true"`
  
  	// An identifier you specify that's included in the completion notification
  	// published to the Amazon SNS topic. For example, you can use JobTag to identify
  	// the type of document that the completion notification corresponds to (such
  	// as a tax form or a receipt).
  	JobTag *string `min:"1" type:"string"`
  
  	// The KMS key used to encrypt the inference results. This can be in either
  	// Key ID or Key Alias format. When a KMS key is provided, the KMS key will
  	// be used for server-side encryption of the objects in the customer bucket.
  	// When this parameter is not enabled, the result will be encrypted server side,using
  	// SSE-S3.
  	KMSKeyId *string `min:"1" type:"string"`
  
  	// The Amazon SNS topic ARN that you want Amazon Textract to publish the completion
  	// status of the operation to.
  	NotificationChannel *NotificationChannel `type:"structure"`
  
  	// Sets if the output will go to a customer defined bucket. By default, Amazon
  	// Textract will save the results internally to be accessed by the GetExpenseAnalysis
  	// operation.
  	OutputConfig *OutputConfig `type:"structure"`
  }

  
  type StartExpenseAnalysisOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the text detection job. The JobId is returned from
  	// StartExpenseAnalysis. A JobId value is only valid for 7 days.
  	JobId *string `min:"1" type:"string"`
  }

  
  type StartLendingAnalysisInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The idempotent token that you use to identify the start request. If you use
  	// the same token with multiple StartLendingAnalysis requests, the same JobId
  	// is returned. Use ClientRequestToken to prevent the same job from being accidentally
  	// started more than once. For more information, see Calling Amazon Textract
  	// Asynchronous Operations (https://docs.aws.amazon.com/textract/latest/dg/api-sync.html).
  	ClientRequestToken *string `min:"1" type:"string"`
  
  	// The Amazon S3 bucket that contains the document to be processed. It's used
  	// by asynchronous operations.
  	//
  	// The input document can be an image file in JPEG or PNG format. It can also
  	// be a file in PDF format.
  	//
  	// DocumentLocation is a required field
  	DocumentLocation *DocumentLocation `type:"structure" required:"true"`
  
  	// An identifier that you specify to be included in the completion notification
  	// published to the Amazon SNS topic. For example, you can use JobTag to identify
  	// the type of document that the completion notification corresponds to (such
  	// as a tax form or a receipt).
  	JobTag *string `min:"1" type:"string"`
  
  	// The KMS key used to encrypt the inference results. This can be in either
  	// Key ID or Key Alias format. When a KMS key is provided, the KMS key will
  	// be used for server-side encryption of the objects in the customer bucket.
  	// When this parameter is not enabled, the result will be encrypted server side,
  	// using SSE-S3.
  	KMSKeyId *string `min:"1" type:"string"`
  
  	// The Amazon Simple Notification Service (Amazon SNS) topic to which Amazon
  	// Textract publishes the completion status of an asynchronous document operation.
  	NotificationChannel *NotificationChannel `type:"structure"`
  
  	// Sets whether or not your output will go to a user created bucket. Used to
  	// set the name of the bucket, and the prefix on the output file.
  	//
  	// OutputConfig is an optional parameter which lets you adjust where your output
  	// will be placed. By default, Amazon Textract will store the results internally
  	// and can only be accessed by the Get API operations. With OutputConfig enabled,
  	// you can set the name of the bucket the output will be sent to the file prefix
  	// of the results where you can download your results. Additionally, you can
  	// set the KMSKeyID parameter to a customer master key (CMK) to encrypt your
  	// output. Without this parameter set Amazon Textract will encrypt server-side
  	// using the AWS managed CMK for Amazon S3.
  	//
  	// Decryption of Customer Content is necessary for processing of the documents
  	// by Amazon Textract. If your account is opted out under an AI services opt
  	// out policy then all unencrypted Customer Content is immediately and permanently
  	// deleted after the Customer Content has been processed by the service. No
  	// copy of of the output is retained by Amazon Textract. For information about
  	// how to opt out, see Managing AI services opt-out policy. (https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_ai-opt-out.html)
  	//
  	// For more information on data privacy, see the Data Privacy FAQ (https://aws.amazon.com/compliance/data-privacy-faq/).
  	OutputConfig *OutputConfig `type:"structure"`
  }

  
  type StartLendingAnalysisOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A unique identifier for the lending or text-detection job. The JobId is returned
  	// from StartLendingAnalysis. A JobId value is only valid for 7 days.
  	JobId *string `min:"1" type:"string"`
  }

  
  type TagResourceInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The Amazon Resource Name (ARN) that specifies the resource to be tagged.
  	//
  	// ResourceARN is a required field
  	ResourceARN *string `min:"1" type:"string" required:"true"`
  
  	// A set of tags (key-value pairs) that you want to assign to the resource.
  	//
  	// Tags is a required field
  	Tags map[string]*string `type:"map" required:"true"`
  }

  
  type TagResourceOutput   struct {
  	_ struct{} `type:"structure"`
  }

  // Amazon Textract is temporarily unable to process the request. Try your call
// again. 
  type ThrottlingException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // A structure containing information about an undetected signature on a page
// where it was expected but not found. 
  type UndetectedSignature   struct {
  	_ struct{} `type:"structure"`
  
  	// The page where a signature was expected but not found.
  	Page *int64 `type:"integer"`
  }

  // The format of the input document isn't supported. Documents for operations
// can be in PNG, JPEG, PDF, or TIFF format. 
  type UnsupportedDocumentException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  
  type UntagResourceInput   struct {
  	_ struct{} `type:"structure"`
  
  	// The Amazon Resource Name (ARN) that specifies the resource to be untagged.
  	//
  	// ResourceARN is a required field
  	ResourceARN *string `min:"1" type:"string" required:"true"`
  
  	// Specifies the tags to be removed from the resource specified by the ResourceARN.
  	//
  	// TagKeys is a required field
  	TagKeys []*string `type:"list" required:"true"`
  }

  
  type UntagResourceOutput   struct {
  	_ struct{} `type:"structure"`
  }

  
  type UpdateAdapterInput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter that will be updated.
  	//
  	// AdapterId is a required field
  	AdapterId *string `min:"12" type:"string" required:"true"`
  
  	// The new name to be applied to the adapter.
  	AdapterName *string `min:"1" type:"string"`
  
  	// The new auto-update status to be applied to the adapter.
  	AutoUpdate *string `type:"string" enum:"AutoUpdate"`
  
  	// The new description to be applied to the adapter.
  	Description *string `min:"1" type:"string"`
  }

  
  type UpdateAdapterOutput   struct {
  	_ struct{} `type:"structure"`
  
  	// A string containing a unique ID for the adapter that has been updated.
  	AdapterId *string `min:"12" type:"string"`
  
  	// A string containing the name of the adapter that has been updated.
  	AdapterName *string `min:"1" type:"string"`
  
  	// The auto-update status of the adapter that has been updated.
  	AutoUpdate *string `type:"string" enum:"AutoUpdate"`
  
  	// An object specifying the creation time of the the adapter that has been updated.
  	CreationTime *time.Time `type:"timestamp"`
  
  	// A string containing the description of the adapter that has been updated.
  	Description *string `min:"1" type:"string"`
  
  	// List of the targeted feature types for the updated adapter.
  	FeatureTypes []*string `type:"list" enum:"FeatureType"`
  }

  // Indicates that a request was not valid. Check request for proper formatting. 
  type ValidationException   struct {
  	_            struct{}                  `type:"structure"`
  	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`
  
  	Message_ *string `locationName:"message" type:"string"`
  }

  // A warning about an issue that occurred during asynchronous text analysis
// (StartDocumentAnalysis) or asynchronous document text detection (StartDocumentTextDetection). 
  type Warning   struct {
  	_ struct{} `type:"structure"`
  
  	// The error code for the warning.
  	ErrorCode *string `type:"string"`
  
  	// A list of the pages that the warning applies to.
  	Pages []*int64 `type:"list"`
  }


Interface Declarations:
