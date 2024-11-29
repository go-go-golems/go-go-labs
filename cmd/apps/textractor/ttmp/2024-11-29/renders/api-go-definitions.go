// You aren't authorized to perform the action. Use the Amazon Resource Name
  // (ARN) of an authorized user or IAM role to perform the operation.
type AccessDeniedException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // An adapter selected for use when analyzing documents. Contains an adapter
  // ID and a version number. Contains information on pages selected for analysis
  // when analyzing documents asychronously.
type Adapter struct {
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
type AdapterOverview struct {
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
type AdapterVersionDatasetConfig struct {
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
type AdapterVersionEvaluationMetric struct {
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
type AdapterVersionOverview struct {
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
type AdaptersConfig struct {
	_ struct{} `type:"structure"`

	// A list of adapters to be used when analyzing the specified document.
	//
	// Adapters is a required field
	Adapters []*Adapter `min:"1" type:"list" required:"true"`
}
type AnalyzeDocumentInput struct {
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
type AnalyzeDocumentOutput struct {
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
type AnalyzeExpenseInput struct {
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
type AnalyzeExpenseOutput struct {
	_ struct{} `type:"structure"`

	// Information about the input document.
	DocumentMetadata *DocumentMetadata `type:"structure"`

	// The expenses detected by Amazon Textract.
	ExpenseDocuments []*ExpenseDocument `type:"list"`
}
  // Used to contain the information detected by an AnalyzeID operation.
type AnalyzeIDDetections struct {
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
type AnalyzeIDInput struct {
	_ struct{} `type:"structure"`

	// The document being passed to AnalyzeID.
	//
	// DocumentPages is a required field
	DocumentPages []*Document `min:"1" type:"list" required:"true"`
}
type AnalyzeIDOutput struct {
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
type BadDocumentException struct {
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
type Block struct {
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
type BoundingBox struct {
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
type ConflictException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
type CreateAdapterInput struct {
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
type CreateAdapterOutput struct {
	_ struct{} `type:"structure"`

	// A string containing the unique ID for the adapter that has been created.
	AdapterId *string `min:"12" type:"string"`
}
type CreateAdapterVersionInput struct {
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
type CreateAdapterVersionOutput struct {
	_ struct{} `type:"structure"`

	// A string containing the unique ID for the adapter that has received a new
	// version.
	AdapterId *string `min:"12" type:"string"`

	// A string describing the new version of the adapter.
	AdapterVersion *string `min:"1" type:"string"`
}
type DeleteAdapterInput struct {
	_ struct{} `type:"structure"`

	// A string containing a unique ID for the adapter to be deleted.
	//
	// AdapterId is a required field
	AdapterId *string `min:"12" type:"string" required:"true"`
}
type DeleteAdapterOutput struct {
	_ struct{} `type:"structure"`
}
type DeleteAdapterVersionInput struct {
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
type DeleteAdapterVersionOutput struct {
	_ struct{} `type:"structure"`
}
type DetectDocumentTextInput struct {
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
type DetectDocumentTextOutput struct {
	_ struct{} `type:"structure"`

	// An array of Block objects that contain the text that's detected in the document.
	Blocks []*Block `type:"list"`

	DetectDocumentTextModelVersion *string `type:"string"`

	// Metadata about the document. It contains the number of pages that are detected
	// in the document.
	DocumentMetadata *DocumentMetadata `type:"structure"`
}
  // A structure that holds information regarding a detected signature on a page.
type DetectedSignature struct {
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
type Document struct {
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
type DocumentGroup struct {
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
type DocumentLocation struct {
	_ struct{} `type:"structure"`

	// The Amazon S3 bucket that contains the input document.
	S3Object *S3Object `type:"structure"`
}
  // Information about the input document.
type DocumentMetadata struct {
	_ struct{} `type:"structure"`

	// The number of pages that are detected in the document.
	Pages *int64 `type:"integer"`
}
  // The document can't be processed because it's too large. The maximum document
  // size for synchronous operations 10 MB. The maximum document size for asynchronous
  // operations is 500 MB for PDF files.
type DocumentTooLargeException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // The evaluation metrics (F1 score, Precision, and Recall) for an adapter version.
type EvaluationMetric struct {
	_ struct{} `type:"structure"`

	// The F1 score for an adapter version.
	F1Score *float64 `type:"float"`

	// The Precision score for an adapter version.
	Precision *float64 `type:"float"`

	// The Recall score for an adapter version.
	Recall *float64 `type:"float"`
}
  // Returns the kind of currency detected.
type ExpenseCurrency struct {
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
type ExpenseDetection struct {
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
type ExpenseDocument struct {
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
type ExpenseField struct {
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
type ExpenseGroupProperty struct {
	_ struct{} `type:"structure"`

	// Provides a group Id number, which will be the same for each in the group.
	Id *string `type:"string"`

	// Informs you on whether the expense group is a name or an address.
	Types []*string `type:"list"`
}
  // An object used to store information about the Type detected by Amazon Textract.
type ExpenseType struct {
	_ struct{} `type:"structure"`

	// The confidence of accuracy, as a percentage.
	Confidence *float64 `type:"float"`

	// The word or line of text detected by Amazon Textract.
	Text *string `type:"string"`
}
  // Contains information extracted by an analysis operation after using StartLendingAnalysis.
type Extraction struct {
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
type Geometry struct {
	_ struct{} `type:"structure"`

	// An axis-aligned coarse representation of the location of the recognized item
	// on the document page.
	BoundingBox *BoundingBox `type:"structure"`

	// Within the bounding box, a fine-grained polygon around the recognized item.
	Polygon []*Point `type:"list"`
}
type GetAdapterInput struct {
	_ struct{} `type:"structure"`

	// A string containing a unique ID for the adapter.
	//
	// AdapterId is a required field
	AdapterId *string `min:"12" type:"string" required:"true"`
}
type GetAdapterOutput struct {
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
type GetAdapterVersionInput struct {
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
type GetAdapterVersionOutput struct {
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
type GetDocumentAnalysisInput struct {
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
type GetDocumentAnalysisOutput struct {
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
type GetDocumentTextDetectionInput struct {
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
type GetDocumentTextDetectionOutput struct {
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
type GetExpenseAnalysisInput struct {
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
type GetExpenseAnalysisOutput struct {
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
type GetLendingAnalysisInput struct {
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
type GetLendingAnalysisOutput struct {
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
type GetLendingAnalysisSummaryInput struct {
	_ struct{} `type:"structure"`

	// A unique identifier for the lending or text-detection job. The JobId is returned
	// from StartLendingAnalysis. A JobId value is only valid for 7 days.
	//
	// JobId is a required field
	JobId *string `min:"1" type:"string" required:"true"`
}
type GetLendingAnalysisSummaryOutput struct {
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
type HumanLoopActivationOutput struct {
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
type HumanLoopConfig struct {
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
type HumanLoopDataAttributes struct {
	_ struct{} `type:"structure"`

	// Sets whether the input image is free of personally identifiable information
	// or adult content.
	ContentClassifiers []*string `type:"list" enum:"ContentClassifier"`
}
  // Indicates you have exceeded the maximum number of active human in the loop
  // workflows available
type HumanLoopQuotaExceededException struct {
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
type IdempotentParameterMismatchException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // The structure that lists each document processed in an AnalyzeID operation.
type IdentityDocument struct {
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
type IdentityDocumentField struct {
	_ struct{} `type:"structure"`

	// Used to contain the information detected by an AnalyzeID operation.
	Type *AnalyzeIDDetections `type:"structure"`

	// Used to contain the information detected by an AnalyzeID operation.
	ValueDetection *AnalyzeIDDetections `type:"structure"`
}
  // Amazon Textract experienced a service issue. Try your call again.
type InternalServerError struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // An invalid job identifier was passed to an asynchronous analysis operation.
type InvalidJobIdException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // Indicates you do not have decrypt permissions with the KMS key entered, or
  // the KMS key was entered incorrectly.
type InvalidKMSKeyException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // An input parameter violated a constraint. For example, in synchronous operations,
  // an InvalidParameterException exception occurs when neither of the S3Object
  // or Bytes values are supplied in the Document request parameter. Validate
  // your parameter before calling the API operation again.
type InvalidParameterException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // Amazon Textract is unable to access the S3 object that's specified in the
  // request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  // For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
type InvalidS3ObjectException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // The results extracted for a lending document.
type LendingDetection struct {
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
type LendingDocument struct {
	_ struct{} `type:"structure"`

	// An array of LendingField objects.
	LendingFields []*LendingField `type:"list"`

	// A list of signatures detected in a lending document.
	SignatureDetections []*SignatureDetection `type:"list"`
}
  // Holds the normalized key-value pairs returned by AnalyzeDocument, including
  // the document type, detected text, and geometry.
type LendingField struct {
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
type LendingResult struct {
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
type LendingSummary struct {
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
type LimitExceededException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // A structure that holds information about the different lines found in a document's
  // tables.
type LineItemFields struct {
	_ struct{} `type:"structure"`

	// ExpenseFields used to show information from detected lines on a table.
	LineItemExpenseFields []*ExpenseField `type:"list"`
}
  // A grouping of tables which contain LineItems, with each table identified
  // by the table's LineItemGroupIndex.
type LineItemGroup struct {
	_ struct{} `type:"structure"`

	// The number used to identify a specific table in a document. The first table
	// encountered will have a LineItemGroupIndex of 1, the second 2, etc.
	LineItemGroupIndex *int64 `type:"integer"`

	// The breakdown of information on a particular line of a table.
	LineItems []*LineItemFields `type:"list"`
}
type ListAdapterVersionsInput struct {
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
type ListAdapterVersionsOutput struct {
	_ struct{} `type:"structure"`

	// Adapter versions that match the filtering criteria specified when calling
	// ListAdapters.
	AdapterVersions []*AdapterVersionOverview `type:"list"`

	// Identifies the next page of results to return when listing adapter versions.
	NextToken *string `min:"1" type:"string"`
}
type ListAdaptersInput struct {
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
type ListAdaptersOutput struct {
	_ struct{} `type:"structure"`

	// A list of adapters that matches the filtering criteria specified when calling
	// ListAdapters.
	Adapters []*AdapterOverview `type:"list"`

	// Identifies the next page of results to return when listing adapters.
	NextToken *string `min:"1" type:"string"`
}
type ListTagsForResourceInput struct {
	_ struct{} `type:"structure"`

	// The Amazon Resource Name (ARN) that specifies the resource to list tags for.
	//
	// ResourceARN is a required field
	ResourceARN *string `min:"1" type:"string" required:"true"`
}
type ListTagsForResourceOutput struct {
	_ struct{} `type:"structure"`

	// A set of tags (key-value pairs) that are part of the requested resource.
	Tags map[string]*string `type:"map"`
}
  // Contains information relating to dates in a document, including the type
  // of value, and the value.
type NormalizedValue struct {
	_ struct{} `type:"structure"`

	// The value of the date, written as Year-Month-DayTHour:Minute:Second.
	Value *string `type:"string"`

	// The normalized type of the value detected. In this case, DATE.
	ValueType *string `type:"string" enum:"ValueType"`
}
  // The Amazon Simple Notification Service (Amazon SNS) topic to which Amazon
  // Textract publishes the completion status of an asynchronous document operation.
type NotificationChannel struct {
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
type OutputConfig struct {
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
type PageClassification struct {
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
type Point struct {
	_ struct{} `type:"structure"`

	// The value of the X coordinate for a point on a Polygon.
	X *float64 `type:"float"`

	// The value of the Y coordinate for a point on a Polygon.
	Y *float64 `type:"float"`
}
  // Contains information regarding predicted values returned by Amazon Textract
  // operations, including the predicted value and the confidence in the predicted
  // value.
type Prediction struct {
	_ struct{} `type:"structure"`

	// Amazon Textract's confidence in its predicted value.
	Confidence *float64 `type:"float"`

	// The predicted value of a detected object.
	Value *string `type:"string"`
}
  // The number of requests exceeded your throughput limit. If you want to increase
  // this limit, contact Amazon Textract.
type ProvisionedThroughputExceededException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
type QueriesConfig struct {
	_ struct{} `type:"structure"`

	// Queries is a required field
	Queries []*Query `min:"1" type:"list" required:"true"`
}
  // Each query contains the question you want to ask in the Text and the alias
  // you want to associate.
type Query struct {
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
type Relationship struct {
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
type ResourceNotFoundException struct {
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
type S3Object struct {
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
type ServiceQuotaExceededException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // Information regarding a detected signature on a page.
type SignatureDetection struct {
	_ struct{} `type:"structure"`

	// The confidence, from 0 to 100, in the predicted values for a detected signature.
	Confidence *float64 `type:"float"`

	// Information about where the following items are located on a document page:
	// detected page, text, key-value pairs, tables, table cells, and selection
	// elements.
	Geometry *Geometry `type:"structure"`
}
  // Contains information about the pages of a document, defined by logical boundary.
type SplitDocument struct {
	_ struct{} `type:"structure"`

	// The index for a given document in a DocumentGroup of a specific Type.
	Index *int64 `type:"integer"`

	// An array of page numbers for a for a given document, ordered by logical boundary.
	Pages []*int64 `type:"list"`
}
type StartDocumentAnalysisInput struct {
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
type StartDocumentAnalysisOutput struct {
	_ struct{} `type:"structure"`

	// The identifier for the document text detection job. Use JobId to identify
	// the job in a subsequent call to GetDocumentAnalysis. A JobId value is only
	// valid for 7 days.
	JobId *string `min:"1" type:"string"`
}
type StartDocumentTextDetectionInput struct {
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
type StartDocumentTextDetectionOutput struct {
	_ struct{} `type:"structure"`

	// The identifier of the text detection job for the document. Use JobId to identify
	// the job in a subsequent call to GetDocumentTextDetection. A JobId value is
	// only valid for 7 days.
	JobId *string `min:"1" type:"string"`
}
type StartExpenseAnalysisInput struct {
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
type StartExpenseAnalysisOutput struct {
	_ struct{} `type:"structure"`

	// A unique identifier for the text detection job. The JobId is returned from
	// StartExpenseAnalysis. A JobId value is only valid for 7 days.
	JobId *string `min:"1" type:"string"`
}
type StartLendingAnalysisInput struct {
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
type StartLendingAnalysisOutput struct {
	_ struct{} `type:"structure"`

	// A unique identifier for the lending or text-detection job. The JobId is returned
	// from StartLendingAnalysis. A JobId value is only valid for 7 days.
	JobId *string `min:"1" type:"string"`
}
type TagResourceInput struct {
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
type TagResourceOutput struct {
	_ struct{} `type:"structure"`
}
  // Amazon Textract is temporarily unable to process the request. Try your call
  // again.
type ThrottlingException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // A structure containing information about an undetected signature on a page
  // where it was expected but not found.
type UndetectedSignature struct {
	_ struct{} `type:"structure"`

	// The page where a signature was expected but not found.
	Page *int64 `type:"integer"`
}
  // The format of the input document isn't supported. Documents for operations
  // can be in PNG, JPEG, PDF, or TIFF format.
type UnsupportedDocumentException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
type UntagResourceInput struct {
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
type UntagResourceOutput struct {
	_ struct{} `type:"structure"`
}
type UpdateAdapterInput struct {
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
type UpdateAdapterOutput struct {
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
type ValidationException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}
  // A warning about an issue that occurred during asynchronous text analysis
  // (StartDocumentAnalysis) or asynchronous document text detection (StartDocumentTextDetection).
type Warning struct {
	_ struct{} `type:"structure"`

	// The error code for the warning.
	ErrorCode *string `type:"string"`

	// A list of the pages that the warning applies to.
	Pages []*int64 `type:"list"`
}

func newErrorAccessDeniedException(v protocol.ResponseMetadata) error 
func newErrorBadDocumentException(v protocol.ResponseMetadata) error 
func newErrorConflictException(v protocol.ResponseMetadata) error 
func newErrorDocumentTooLargeException(v protocol.ResponseMetadata) error 
func newErrorHumanLoopQuotaExceededException(v protocol.ResponseMetadata) error 
func newErrorIdempotentParameterMismatchException(v protocol.ResponseMetadata) error 
func newErrorInternalServerError(v protocol.ResponseMetadata) error 
func newErrorInvalidJobIdException(v protocol.ResponseMetadata) error 
func newErrorInvalidKMSKeyException(v protocol.ResponseMetadata) error 
func newErrorInvalidParameterException(v protocol.ResponseMetadata) error 
func newErrorInvalidS3ObjectException(v protocol.ResponseMetadata) error 
func newErrorLimitExceededException(v protocol.ResponseMetadata) error 
func newErrorProvisionedThroughputExceededException(v protocol.ResponseMetadata) error 
func newErrorResourceNotFoundException(v protocol.ResponseMetadata) error 
func newErrorServiceQuotaExceededException(v protocol.ResponseMetadata) error 
func newErrorThrottlingException(v protocol.ResponseMetadata) error 
func newErrorUnsupportedDocumentException(v protocol.ResponseMetadata) error 
func newErrorValidationException(v protocol.ResponseMetadata) error 
  // AdapterVersionStatus_Values returns all elements of the AdapterVersionStatus enum
func AdapterVersionStatus_Values() []string 
  // AutoUpdate_Values returns all elements of the AutoUpdate enum
func AutoUpdate_Values() []string 
  // BlockType_Values returns all elements of the BlockType enum
func BlockType_Values() []string 
  // ContentClassifier_Values returns all elements of the ContentClassifier enum
func ContentClassifier_Values() []string 
  // EntityType_Values returns all elements of the EntityType enum
func EntityType_Values() []string 
  // FeatureType_Values returns all elements of the FeatureType enum
func FeatureType_Values() []string 
  // JobStatus_Values returns all elements of the JobStatus enum
func JobStatus_Values() []string 
  // RelationshipType_Values returns all elements of the RelationshipType enum
func RelationshipType_Values() []string 
  // SelectionStatus_Values returns all elements of the SelectionStatus enum
func SelectionStatus_Values() []string 
  // TextType_Values returns all elements of the TextType enum
func TextType_Values() []string 
  // ValueType_Values returns all elements of the ValueType enum
func ValueType_Values() []string 
  // AnalyzeDocumentRequest generates a "aws/request.Request" representing the
  // client's request for the AnalyzeDocument operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See AnalyzeDocument for more information on using the AnalyzeDocument
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the AnalyzeDocumentRequest method.
  //	req, resp := client.AnalyzeDocumentRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/AnalyzeDocument
func (c *Textract) AnalyzeDocumentRequest(input *AnalyzeDocumentInput) (req *request.Request, output *AnalyzeDocumentOutput)
  // AnalyzeDocument API operation for Amazon Textract.
  //
  // Analyzes an input document for relationships between detected items.
  //
  // The types of information returned are as follows:
  //
  //   - Form data (key-value pairs). The related information is returned in
  //     two Block objects, each of type KEY_VALUE_SET: a KEY Block object and
  //     a VALUE Block object. For example, Name: Ana Silva Carolina contains a
  //     key and value. Name: is the key. Ana Silva Carolina is the value.
  //
  //   - Table and table cell data. A TABLE Block object contains information
  //     about a detected table. A CELL Block object is returned for each cell
  //     in a table.
  //
  //   - Lines and words of text. A LINE Block object contains one or more WORD
  //     Block objects. All lines and words that are detected in the document are
  //     returned (including text that doesn't have a relationship with the value
  //     of FeatureTypes).
  //
  //   - Signatures. A SIGNATURE Block object contains the location information
  //     of a signature in a document. If used in conjunction with forms or tables,
  //     a signature can be given a Key-Value pairing or be detected in the cell
  //     of a table.
  //
  //   - Query. A QUERY Block object contains the query text, alias and link
  //     to the associated Query results block object.
  //
  //   - Query Result. A QUERY_RESULT Block object contains the answer to the
  //     query and an ID that connects it to the query asked. This Block also contains
  //     a confidence score.
  //
  // Selection elements such as check boxes and option buttons (radio buttons)
  // can be detected in form data and in tables. A SELECTION_ELEMENT Block object
  // contains information about a selection element, including the selection status.
  //
  // You can choose which type of analysis to perform by specifying the FeatureTypes
  // list.
  //
  // The output is returned in a list of Block objects.
  //
  // AnalyzeDocument is a synchronous operation. To analyze documents asynchronously,
  // use StartDocumentAnalysis.
  //
  // For more information, see Document Text Analysis (https://docs.aws.amazon.com/textract/latest/dg/how-it-works-analyzing.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation AnalyzeDocument for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - HumanLoopQuotaExceededException
  //     Indicates you have exceeded the maximum number of active human in the loop
  //     workflows available
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/AnalyzeDocument
func (c *Textract) AnalyzeDocument(input *AnalyzeDocumentInput) (*AnalyzeDocumentOutput, error)
  // AnalyzeDocumentWithContext is the same as AnalyzeDocument with the addition of
  // the ability to pass a context and additional request options.
  //
  // See AnalyzeDocument for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) AnalyzeDocumentWithContext(ctx aws.Context, input *AnalyzeDocumentInput, opts ...request.Option) (*AnalyzeDocumentOutput, error)
  // AnalyzeExpenseRequest generates a "aws/request.Request" representing the
  // client's request for the AnalyzeExpense operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See AnalyzeExpense for more information on using the AnalyzeExpense
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the AnalyzeExpenseRequest method.
  //	req, resp := client.AnalyzeExpenseRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/AnalyzeExpense
func (c *Textract) AnalyzeExpenseRequest(input *AnalyzeExpenseInput) (req *request.Request, output *AnalyzeExpenseOutput)
  // AnalyzeExpense API operation for Amazon Textract.
  //
  // AnalyzeExpense synchronously analyzes an input document for financially related
  // relationships between text.
  //
  // Information is returned as ExpenseDocuments and seperated as follows:
  //
  //   - LineItemGroups- A data set containing LineItems which store information
  //     about the lines of text, such as an item purchased and its price on a
  //     receipt.
  //
  //   - SummaryFields- Contains all other information a receipt, such as header
  //     information or the vendors name.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation AnalyzeExpense for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/AnalyzeExpense
func (c *Textract) AnalyzeExpense(input *AnalyzeExpenseInput) (*AnalyzeExpenseOutput, error)
  // AnalyzeExpenseWithContext is the same as AnalyzeExpense with the addition of
  // the ability to pass a context and additional request options.
  //
  // See AnalyzeExpense for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) AnalyzeExpenseWithContext(ctx aws.Context, input *AnalyzeExpenseInput, opts ...request.Option) (*AnalyzeExpenseOutput, error)
  // AnalyzeIDRequest generates a "aws/request.Request" representing the
  // client's request for the AnalyzeID operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See AnalyzeID for more information on using the AnalyzeID
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the AnalyzeIDRequest method.
  //	req, resp := client.AnalyzeIDRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/AnalyzeID
func (c *Textract) AnalyzeIDRequest(input *AnalyzeIDInput) (req *request.Request, output *AnalyzeIDOutput)
  // AnalyzeID API operation for Amazon Textract.
  //
  // Analyzes identity documents for relevant information. This information is
  // extracted and returned as IdentityDocumentFields, which records both the
  // normalized field and value of the extracted text. Unlike other Amazon Textract
  // operations, AnalyzeID doesn't return any Geometry data.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation AnalyzeID for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/AnalyzeID
func (c *Textract) AnalyzeID(input *AnalyzeIDInput) (*AnalyzeIDOutput, error)
  // AnalyzeIDWithContext is the same as AnalyzeID with the addition of
  // the ability to pass a context and additional request options.
  //
  // See AnalyzeID for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) AnalyzeIDWithContext(ctx aws.Context, input *AnalyzeIDInput, opts ...request.Option) (*AnalyzeIDOutput, error)
  // CreateAdapterRequest generates a "aws/request.Request" representing the
  // client's request for the CreateAdapter operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See CreateAdapter for more information on using the CreateAdapter
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the CreateAdapterRequest method.
  //	req, resp := client.CreateAdapterRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/CreateAdapter
func (c *Textract) CreateAdapterRequest(input *CreateAdapterInput) (req *request.Request, output *CreateAdapterOutput)
  // CreateAdapter API operation for Amazon Textract.
  //
  // Creates an adapter, which can be fine-tuned for enhanced performance on user
  // provided documents. Takes an AdapterName and FeatureType. Currently the only
  // supported feature type is QUERIES. You can also provide a Description, Tags,
  // and a ClientRequestToken. You can choose whether or not the adapter should
  // be AutoUpdated with the AutoUpdate argument. By default, AutoUpdate is set
  // to DISABLED.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation CreateAdapter for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ConflictException
  //     Updating or deleting a resource can cause an inconsistent state.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - IdempotentParameterMismatchException
  //     A ClientRequestToken input parameter was reused with an operation, but at
  //     least one of the other input parameters is different from the previous call
  //     to the operation.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - LimitExceededException
  //     An Amazon Textract service limit was exceeded. For example, if you start
  //     too many asynchronous jobs concurrently, calls to start operations (StartDocumentTextDetection,
  //     for example) raise a LimitExceededException exception (HTTP status code:
  //     400) until the number of concurrently running jobs is below the Amazon Textract
  //     service limit.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ServiceQuotaExceededException
  //     Returned when a request cannot be completed as it would exceed a maximum
  //     service quota.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/CreateAdapter
func (c *Textract) CreateAdapter(input *CreateAdapterInput) (*CreateAdapterOutput, error)
  // CreateAdapterWithContext is the same as CreateAdapter with the addition of
  // the ability to pass a context and additional request options.
  //
  // See CreateAdapter for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) CreateAdapterWithContext(ctx aws.Context, input *CreateAdapterInput, opts ...request.Option) (*CreateAdapterOutput, error)
  // CreateAdapterVersionRequest generates a "aws/request.Request" representing the
  // client's request for the CreateAdapterVersion operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See CreateAdapterVersion for more information on using the CreateAdapterVersion
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the CreateAdapterVersionRequest method.
  //	req, resp := client.CreateAdapterVersionRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/CreateAdapterVersion
func (c *Textract) CreateAdapterVersionRequest(input *CreateAdapterVersionInput) (req *request.Request, output *CreateAdapterVersionOutput)
  // CreateAdapterVersion API operation for Amazon Textract.
  //
  // Creates a new version of an adapter. Operates on a provided AdapterId and
  // a specified dataset provided via the DatasetConfig argument. Requires that
  // you specify an Amazon S3 bucket with the OutputConfig argument. You can provide
  // an optional KMSKeyId, an optional ClientRequestToken, and optional tags.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation CreateAdapterVersion for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - IdempotentParameterMismatchException
  //     A ClientRequestToken input parameter was reused with an operation, but at
  //     least one of the other input parameters is different from the previous call
  //     to the operation.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - LimitExceededException
  //     An Amazon Textract service limit was exceeded. For example, if you start
  //     too many asynchronous jobs concurrently, calls to start operations (StartDocumentTextDetection,
  //     for example) raise a LimitExceededException exception (HTTP status code:
  //     400) until the number of concurrently running jobs is below the Amazon Textract
  //     service limit.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ServiceQuotaExceededException
  //     Returned when a request cannot be completed as it would exceed a maximum
  //     service quota.
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  //   - ConflictException
  //     Updating or deleting a resource can cause an inconsistent state.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/CreateAdapterVersion
func (c *Textract) CreateAdapterVersion(input *CreateAdapterVersionInput) (*CreateAdapterVersionOutput, error)
  // CreateAdapterVersionWithContext is the same as CreateAdapterVersion with the addition of
  // the ability to pass a context and additional request options.
  //
  // See CreateAdapterVersion for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) CreateAdapterVersionWithContext(ctx aws.Context, input *CreateAdapterVersionInput, opts ...request.Option) (*CreateAdapterVersionOutput, error)
  // DeleteAdapterRequest generates a "aws/request.Request" representing the
  // client's request for the DeleteAdapter operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See DeleteAdapter for more information on using the DeleteAdapter
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the DeleteAdapterRequest method.
  //	req, resp := client.DeleteAdapterRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/DeleteAdapter
func (c *Textract) DeleteAdapterRequest(input *DeleteAdapterInput) (req *request.Request, output *DeleteAdapterOutput)
  // DeleteAdapter API operation for Amazon Textract.
  //
  // Deletes an Amazon Textract adapter. Takes an AdapterId and deletes the adapter
  // specified by the ID.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation DeleteAdapter for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ConflictException
  //     Updating or deleting a resource can cause an inconsistent state.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/DeleteAdapter
func (c *Textract) DeleteAdapter(input *DeleteAdapterInput) (*DeleteAdapterOutput, error)
  // DeleteAdapterWithContext is the same as DeleteAdapter with the addition of
  // the ability to pass a context and additional request options.
  //
  // See DeleteAdapter for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) DeleteAdapterWithContext(ctx aws.Context, input *DeleteAdapterInput, opts ...request.Option) (*DeleteAdapterOutput, error)
  // DeleteAdapterVersionRequest generates a "aws/request.Request" representing the
  // client's request for the DeleteAdapterVersion operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See DeleteAdapterVersion for more information on using the DeleteAdapterVersion
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the DeleteAdapterVersionRequest method.
  //	req, resp := client.DeleteAdapterVersionRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/DeleteAdapterVersion
func (c *Textract) DeleteAdapterVersionRequest(input *DeleteAdapterVersionInput) (req *request.Request, output *DeleteAdapterVersionOutput)
  // DeleteAdapterVersion API operation for Amazon Textract.
  //
  // Deletes an Amazon Textract adapter version. Requires that you specify both
  // an AdapterId and a AdapterVersion. Deletes the adapter version specified
  // by the AdapterId and the AdapterVersion.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation DeleteAdapterVersion for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ConflictException
  //     Updating or deleting a resource can cause an inconsistent state.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/DeleteAdapterVersion
func (c *Textract) DeleteAdapterVersion(input *DeleteAdapterVersionInput) (*DeleteAdapterVersionOutput, error)
  // DeleteAdapterVersionWithContext is the same as DeleteAdapterVersion with the addition of
  // the ability to pass a context and additional request options.
  //
  // See DeleteAdapterVersion for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) DeleteAdapterVersionWithContext(ctx aws.Context, input *DeleteAdapterVersionInput, opts ...request.Option) (*DeleteAdapterVersionOutput, error)
  // DetectDocumentTextRequest generates a "aws/request.Request" representing the
  // client's request for the DetectDocumentText operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See DetectDocumentText for more information on using the DetectDocumentText
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the DetectDocumentTextRequest method.
  //	req, resp := client.DetectDocumentTextRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/DetectDocumentText
func (c *Textract) DetectDocumentTextRequest(input *DetectDocumentTextInput) (req *request.Request, output *DetectDocumentTextOutput)
  // DetectDocumentText API operation for Amazon Textract.
  //
  // Detects text in the input document. Amazon Textract can detect lines of text
  // and the words that make up a line of text. The input document must be in
  // one of the following image formats: JPEG, PNG, PDF, or TIFF. DetectDocumentText
  // returns the detected text in an array of Block objects.
  //
  // Each document page has as an associated Block of type PAGE. Each PAGE Block
  // object is the parent of LINE Block objects that represent the lines of detected
  // text on a page. A LINE Block object is a parent for each word that makes
  // up the line. Words are represented by Block objects of type WORD.
  //
  // DetectDocumentText is a synchronous operation. To analyze documents asynchronously,
  // use StartDocumentTextDetection.
  //
  // For more information, see Document Text Detection (https://docs.aws.amazon.com/textract/latest/dg/how-it-works-detecting.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation DetectDocumentText for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/DetectDocumentText
func (c *Textract) DetectDocumentText(input *DetectDocumentTextInput) (*DetectDocumentTextOutput, error)
  // DetectDocumentTextWithContext is the same as DetectDocumentText with the addition of
  // the ability to pass a context and additional request options.
  //
  // See DetectDocumentText for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) DetectDocumentTextWithContext(ctx aws.Context, input *DetectDocumentTextInput, opts ...request.Option) (*DetectDocumentTextOutput, error)
  // GetAdapterRequest generates a "aws/request.Request" representing the
  // client's request for the GetAdapter operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See GetAdapter for more information on using the GetAdapter
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the GetAdapterRequest method.
  //	req, resp := client.GetAdapterRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetAdapter
func (c *Textract) GetAdapterRequest(input *GetAdapterInput) (req *request.Request, output *GetAdapterOutput)
  // GetAdapter API operation for Amazon Textract.
  //
  // Gets configuration information for an adapter specified by an AdapterId,
  // returning information on AdapterName, Description, CreationTime, AutoUpdate
  // status, and FeatureTypes.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation GetAdapter for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetAdapter
func (c *Textract) GetAdapter(input *GetAdapterInput) (*GetAdapterOutput, error)
  // GetAdapterWithContext is the same as GetAdapter with the addition of
  // the ability to pass a context and additional request options.
  //
  // See GetAdapter for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) GetAdapterWithContext(ctx aws.Context, input *GetAdapterInput, opts ...request.Option) (*GetAdapterOutput, error)
  // GetAdapterVersionRequest generates a "aws/request.Request" representing the
  // client's request for the GetAdapterVersion operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See GetAdapterVersion for more information on using the GetAdapterVersion
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the GetAdapterVersionRequest method.
  //	req, resp := client.GetAdapterVersionRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetAdapterVersion
func (c *Textract) GetAdapterVersionRequest(input *GetAdapterVersionInput) (req *request.Request, output *GetAdapterVersionOutput)
  // GetAdapterVersion API operation for Amazon Textract.
  //
  // Gets configuration information for the specified adapter version, including:
  // AdapterId, AdapterVersion, FeatureTypes, Status, StatusMessage, DatasetConfig,
  // KMSKeyId, OutputConfig, Tags and EvaluationMetrics.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation GetAdapterVersion for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetAdapterVersion
func (c *Textract) GetAdapterVersion(input *GetAdapterVersionInput) (*GetAdapterVersionOutput, error)
  // GetAdapterVersionWithContext is the same as GetAdapterVersion with the addition of
  // the ability to pass a context and additional request options.
  //
  // See GetAdapterVersion for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) GetAdapterVersionWithContext(ctx aws.Context, input *GetAdapterVersionInput, opts ...request.Option) (*GetAdapterVersionOutput, error)
  // GetDocumentAnalysisRequest generates a "aws/request.Request" representing the
  // client's request for the GetDocumentAnalysis operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See GetDocumentAnalysis for more information on using the GetDocumentAnalysis
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the GetDocumentAnalysisRequest method.
  //	req, resp := client.GetDocumentAnalysisRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetDocumentAnalysis
func (c *Textract) GetDocumentAnalysisRequest(input *GetDocumentAnalysisInput) (req *request.Request, output *GetDocumentAnalysisOutput)
  // GetDocumentAnalysis API operation for Amazon Textract.
  //
  // Gets the results for an Amazon Textract asynchronous operation that analyzes
  // text in a document.
  //
  // You start asynchronous text analysis by calling StartDocumentAnalysis, which
  // returns a job identifier (JobId). When the text analysis operation finishes,
  // Amazon Textract publishes a completion status to the Amazon Simple Notification
  // Service (Amazon SNS) topic that's registered in the initial call to StartDocumentAnalysis.
  // To get the results of the text-detection operation, first check that the
  // status value published to the Amazon SNS topic is SUCCEEDED. If so, call
  // GetDocumentAnalysis, and pass the job identifier (JobId) from the initial
  // call to StartDocumentAnalysis.
  //
  // GetDocumentAnalysis returns an array of Block objects. The following types
  // of information are returned:
  //
  //   - Form data (key-value pairs). The related information is returned in
  //     two Block objects, each of type KEY_VALUE_SET: a KEY Block object and
  //     a VALUE Block object. For example, Name: Ana Silva Carolina contains a
  //     key and value. Name: is the key. Ana Silva Carolina is the value.
  //
  //   - Table and table cell data. A TABLE Block object contains information
  //     about a detected table. A CELL Block object is returned for each cell
  //     in a table.
  //
  //   - Lines and words of text. A LINE Block object contains one or more WORD
  //     Block objects. All lines and words that are detected in the document are
  //     returned (including text that doesn't have a relationship with the value
  //     of the StartDocumentAnalysis FeatureTypes input parameter).
  //
  //   - Query. A QUERY Block object contains the query text, alias and link
  //     to the associated Query results block object.
  //
  //   - Query Results. A QUERY_RESULT Block object contains the answer to the
  //     query and an ID that connects it to the query asked. This Block also contains
  //     a confidence score.
  //
  // While processing a document with queries, look out for INVALID_REQUEST_PARAMETERS
  // output. This indicates that either the per page query limit has been exceeded
  // or that the operation is trying to query a page in the document which doesn’t
  // exist.
  //
  // Selection elements such as check boxes and option buttons (radio buttons)
  // can be detected in form data and in tables. A SELECTION_ELEMENT Block object
  // contains information about a selection element, including the selection status.
  //
  // Use the MaxResults parameter to limit the number of blocks that are returned.
  // If there are more results than specified in MaxResults, the value of NextToken
  // in the operation response contains a pagination token for getting the next
  // set of results. To get the next page of results, call GetDocumentAnalysis,
  // and populate the NextToken request parameter with the token value that's
  // returned from the previous call to GetDocumentAnalysis.
  //
  // For more information, see Document Text Analysis (https://docs.aws.amazon.com/textract/latest/dg/how-it-works-analyzing.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation GetDocumentAnalysis for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InvalidJobIdException
  //     An invalid job identifier was passed to an asynchronous analysis operation.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetDocumentAnalysis
func (c *Textract) GetDocumentAnalysis(input *GetDocumentAnalysisInput) (*GetDocumentAnalysisOutput, error)
  // GetDocumentAnalysisWithContext is the same as GetDocumentAnalysis with the addition of
  // the ability to pass a context and additional request options.
  //
  // See GetDocumentAnalysis for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) GetDocumentAnalysisWithContext(ctx aws.Context, input *GetDocumentAnalysisInput, opts ...request.Option) (*GetDocumentAnalysisOutput, error)
  // GetDocumentTextDetectionRequest generates a "aws/request.Request" representing the
  // client's request for the GetDocumentTextDetection operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See GetDocumentTextDetection for more information on using the GetDocumentTextDetection
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the GetDocumentTextDetectionRequest method.
  //	req, resp := client.GetDocumentTextDetectionRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetDocumentTextDetection
func (c *Textract) GetDocumentTextDetectionRequest(input *GetDocumentTextDetectionInput) (req *request.Request, output *GetDocumentTextDetectionOutput)
  // GetDocumentTextDetection API operation for Amazon Textract.
  //
  // Gets the results for an Amazon Textract asynchronous operation that detects
  // text in a document. Amazon Textract can detect lines of text and the words
  // that make up a line of text.
  //
  // You start asynchronous text detection by calling StartDocumentTextDetection,
  // which returns a job identifier (JobId). When the text detection operation
  // finishes, Amazon Textract publishes a completion status to the Amazon Simple
  // Notification Service (Amazon SNS) topic that's registered in the initial
  // call to StartDocumentTextDetection. To get the results of the text-detection
  // operation, first check that the status value published to the Amazon SNS
  // topic is SUCCEEDED. If so, call GetDocumentTextDetection, and pass the job
  // identifier (JobId) from the initial call to StartDocumentTextDetection.
  //
  // GetDocumentTextDetection returns an array of Block objects.
  //
  // Each document page has as an associated Block of type PAGE. Each PAGE Block
  // object is the parent of LINE Block objects that represent the lines of detected
  // text on a page. A LINE Block object is a parent for each word that makes
  // up the line. Words are represented by Block objects of type WORD.
  //
  // Use the MaxResults parameter to limit the number of blocks that are returned.
  // If there are more results than specified in MaxResults, the value of NextToken
  // in the operation response contains a pagination token for getting the next
  // set of results. To get the next page of results, call GetDocumentTextDetection,
  // and populate the NextToken request parameter with the token value that's
  // returned from the previous call to GetDocumentTextDetection.
  //
  // For more information, see Document Text Detection (https://docs.aws.amazon.com/textract/latest/dg/how-it-works-detecting.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation GetDocumentTextDetection for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InvalidJobIdException
  //     An invalid job identifier was passed to an asynchronous analysis operation.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetDocumentTextDetection
func (c *Textract) GetDocumentTextDetection(input *GetDocumentTextDetectionInput) (*GetDocumentTextDetectionOutput, error)
  // GetDocumentTextDetectionWithContext is the same as GetDocumentTextDetection with the addition of
  // the ability to pass a context and additional request options.
  //
  // See GetDocumentTextDetection for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) GetDocumentTextDetectionWithContext(ctx aws.Context, input *GetDocumentTextDetectionInput, opts ...request.Option) (*GetDocumentTextDetectionOutput, error)
  // GetExpenseAnalysisRequest generates a "aws/request.Request" representing the
  // client's request for the GetExpenseAnalysis operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See GetExpenseAnalysis for more information on using the GetExpenseAnalysis
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the GetExpenseAnalysisRequest method.
  //	req, resp := client.GetExpenseAnalysisRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetExpenseAnalysis
func (c *Textract) GetExpenseAnalysisRequest(input *GetExpenseAnalysisInput) (req *request.Request, output *GetExpenseAnalysisOutput)
  // GetExpenseAnalysis API operation for Amazon Textract.
  //
  // Gets the results for an Amazon Textract asynchronous operation that analyzes
  // invoices and receipts. Amazon Textract finds contact information, items purchased,
  // and vendor name, from input invoices and receipts.
  //
  // You start asynchronous invoice/receipt analysis by calling StartExpenseAnalysis,
  // which returns a job identifier (JobId). Upon completion of the invoice/receipt
  // analysis, Amazon Textract publishes the completion status to the Amazon Simple
  // Notification Service (Amazon SNS) topic. This topic must be registered in
  // the initial call to StartExpenseAnalysis. To get the results of the invoice/receipt
  // analysis operation, first ensure that the status value published to the Amazon
  // SNS topic is SUCCEEDED. If so, call GetExpenseAnalysis, and pass the job
  // identifier (JobId) from the initial call to StartExpenseAnalysis.
  //
  // Use the MaxResults parameter to limit the number of blocks that are returned.
  // If there are more results than specified in MaxResults, the value of NextToken
  // in the operation response contains a pagination token for getting the next
  // set of results. To get the next page of results, call GetExpenseAnalysis,
  // and populate the NextToken request parameter with the token value that's
  // returned from the previous call to GetExpenseAnalysis.
  //
  // For more information, see Analyzing Invoices and Receipts (https://docs.aws.amazon.com/textract/latest/dg/invoices-receipts.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation GetExpenseAnalysis for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InvalidJobIdException
  //     An invalid job identifier was passed to an asynchronous analysis operation.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetExpenseAnalysis
func (c *Textract) GetExpenseAnalysis(input *GetExpenseAnalysisInput) (*GetExpenseAnalysisOutput, error)
  // GetExpenseAnalysisWithContext is the same as GetExpenseAnalysis with the addition of
  // the ability to pass a context and additional request options.
  //
  // See GetExpenseAnalysis for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) GetExpenseAnalysisWithContext(ctx aws.Context, input *GetExpenseAnalysisInput, opts ...request.Option) (*GetExpenseAnalysisOutput, error)
  // GetLendingAnalysisRequest generates a "aws/request.Request" representing the
  // client's request for the GetLendingAnalysis operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See GetLendingAnalysis for more information on using the GetLendingAnalysis
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the GetLendingAnalysisRequest method.
  //	req, resp := client.GetLendingAnalysisRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetLendingAnalysis
func (c *Textract) GetLendingAnalysisRequest(input *GetLendingAnalysisInput) (req *request.Request, output *GetLendingAnalysisOutput)
  // GetLendingAnalysis API operation for Amazon Textract.
  //
  // Gets the results for an Amazon Textract asynchronous operation that analyzes
  // text in a lending document.
  //
  // You start asynchronous text analysis by calling StartLendingAnalysis, which
  // returns a job identifier (JobId). When the text analysis operation finishes,
  // Amazon Textract publishes a completion status to the Amazon Simple Notification
  // Service (Amazon SNS) topic that's registered in the initial call to StartLendingAnalysis.
  //
  // To get the results of the text analysis operation, first check that the status
  // value published to the Amazon SNS topic is SUCCEEDED. If so, call GetLendingAnalysis,
  // and pass the job identifier (JobId) from the initial call to StartLendingAnalysis.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation GetLendingAnalysis for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InvalidJobIdException
  //     An invalid job identifier was passed to an asynchronous analysis operation.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetLendingAnalysis
func (c *Textract) GetLendingAnalysis(input *GetLendingAnalysisInput) (*GetLendingAnalysisOutput, error)
  // GetLendingAnalysisWithContext is the same as GetLendingAnalysis with the addition of
  // the ability to pass a context and additional request options.
  //
  // See GetLendingAnalysis for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) GetLendingAnalysisWithContext(ctx aws.Context, input *GetLendingAnalysisInput, opts ...request.Option) (*GetLendingAnalysisOutput, error)
  // GetLendingAnalysisSummaryRequest generates a "aws/request.Request" representing the
  // client's request for the GetLendingAnalysisSummary operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See GetLendingAnalysisSummary for more information on using the GetLendingAnalysisSummary
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the GetLendingAnalysisSummaryRequest method.
  //	req, resp := client.GetLendingAnalysisSummaryRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetLendingAnalysisSummary
func (c *Textract) GetLendingAnalysisSummaryRequest(input *GetLendingAnalysisSummaryInput) (req *request.Request, output *GetLendingAnalysisSummaryOutput)
  // GetLendingAnalysisSummary API operation for Amazon Textract.
  //
  // Gets summarized results for the StartLendingAnalysis operation, which analyzes
  // text in a lending document. The returned summary consists of information
  // about documents grouped together by a common document type. Information like
  // detected signatures, page numbers, and split documents is returned with respect
  // to the type of grouped document.
  //
  // You start asynchronous text analysis by calling StartLendingAnalysis, which
  // returns a job identifier (JobId). When the text analysis operation finishes,
  // Amazon Textract publishes a completion status to the Amazon Simple Notification
  // Service (Amazon SNS) topic that's registered in the initial call to StartLendingAnalysis.
  //
  // To get the results of the text analysis operation, first check that the status
  // value published to the Amazon SNS topic is SUCCEEDED. If so, call GetLendingAnalysisSummary,
  // and pass the job identifier (JobId) from the initial call to StartLendingAnalysis.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation GetLendingAnalysisSummary for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InvalidJobIdException
  //     An invalid job identifier was passed to an asynchronous analysis operation.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/GetLendingAnalysisSummary
func (c *Textract) GetLendingAnalysisSummary(input *GetLendingAnalysisSummaryInput) (*GetLendingAnalysisSummaryOutput, error)
  // GetLendingAnalysisSummaryWithContext is the same as GetLendingAnalysisSummary with the addition of
  // the ability to pass a context and additional request options.
  //
  // See GetLendingAnalysisSummary for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) GetLendingAnalysisSummaryWithContext(ctx aws.Context, input *GetLendingAnalysisSummaryInput, opts ...request.Option) (*GetLendingAnalysisSummaryOutput, error)
  // ListAdapterVersionsRequest generates a "aws/request.Request" representing the
  // client's request for the ListAdapterVersions operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See ListAdapterVersions for more information on using the ListAdapterVersions
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the ListAdapterVersionsRequest method.
  //	req, resp := client.ListAdapterVersionsRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/ListAdapterVersions
func (c *Textract) ListAdapterVersionsRequest(input *ListAdapterVersionsInput) (req *request.Request, output *ListAdapterVersionsOutput)
  // ListAdapterVersions API operation for Amazon Textract.
  //
  // List all version of an adapter that meet the specified filtration criteria.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation ListAdapterVersions for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/ListAdapterVersions
func (c *Textract) ListAdapterVersions(input *ListAdapterVersionsInput) (*ListAdapterVersionsOutput, error)
  // ListAdapterVersionsWithContext is the same as ListAdapterVersions with the addition of
  // the ability to pass a context and additional request options.
  //
  // See ListAdapterVersions for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) ListAdapterVersionsWithContext(ctx aws.Context, input *ListAdapterVersionsInput, opts ...request.Option) (*ListAdapterVersionsOutput, error)
  // ListAdapterVersionsPages iterates over the pages of a ListAdapterVersions operation,
  // calling the "fn" function with the response data for each page. To stop
  // iterating, return false from the fn function.
  //
  // See ListAdapterVersions method for more information on how to use this operation.
  //
  // Note: This operation can generate multiple requests to a service.
  //
  //	// Example iterating over at most 3 pages of a ListAdapterVersions operation.
  //	pageNum := 0
  //	err := client.ListAdapterVersionsPages(params,
  //	    func(page *textract.ListAdapterVersionsOutput, lastPage bool) bool {
  //	        pageNum++
  //	        fmt.Println(page)
  //	        return pageNum <= 3
  //	    })
func (c *Textract) ListAdapterVersionsPages(input *ListAdapterVersionsInput, fn func(*ListAdapterVersionsOutput, bool) bool) error
  // ListAdapterVersionsPagesWithContext same as ListAdapterVersionsPages except
  // it takes a Context and allows setting request options on the pages.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) ListAdapterVersionsPagesWithContext(ctx aws.Context, input *ListAdapterVersionsInput, fn func(*ListAdapterVersionsOutput, bool) bool, opts ...request.Option) error
  // ListAdaptersRequest generates a "aws/request.Request" representing the
  // client's request for the ListAdapters operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See ListAdapters for more information on using the ListAdapters
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the ListAdaptersRequest method.
  //	req, resp := client.ListAdaptersRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/ListAdapters
func (c *Textract) ListAdaptersRequest(input *ListAdaptersInput) (req *request.Request, output *ListAdaptersOutput)
  // ListAdapters API operation for Amazon Textract.
  //
  // Lists all adapters that match the specified filtration criteria.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation ListAdapters for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/ListAdapters
func (c *Textract) ListAdapters(input *ListAdaptersInput) (*ListAdaptersOutput, error)
  // ListAdaptersWithContext is the same as ListAdapters with the addition of
  // the ability to pass a context and additional request options.
  //
  // See ListAdapters for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) ListAdaptersWithContext(ctx aws.Context, input *ListAdaptersInput, opts ...request.Option) (*ListAdaptersOutput, error)
  // ListAdaptersPages iterates over the pages of a ListAdapters operation,
  // calling the "fn" function with the response data for each page. To stop
  // iterating, return false from the fn function.
  //
  // See ListAdapters method for more information on how to use this operation.
  //
  // Note: This operation can generate multiple requests to a service.
  //
  //	// Example iterating over at most 3 pages of a ListAdapters operation.
  //	pageNum := 0
  //	err := client.ListAdaptersPages(params,
  //	    func(page *textract.ListAdaptersOutput, lastPage bool) bool {
  //	        pageNum++
  //	        fmt.Println(page)
  //	        return pageNum <= 3
  //	    })
func (c *Textract) ListAdaptersPages(input *ListAdaptersInput, fn func(*ListAdaptersOutput, bool) bool) error
  // ListAdaptersPagesWithContext same as ListAdaptersPages except
  // it takes a Context and allows setting request options on the pages.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) ListAdaptersPagesWithContext(ctx aws.Context, input *ListAdaptersInput, fn func(*ListAdaptersOutput, bool) bool, opts ...request.Option) error
  // ListTagsForResourceRequest generates a "aws/request.Request" representing the
  // client's request for the ListTagsForResource operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See ListTagsForResource for more information on using the ListTagsForResource
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the ListTagsForResourceRequest method.
  //	req, resp := client.ListTagsForResourceRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/ListTagsForResource
func (c *Textract) ListTagsForResourceRequest(input *ListTagsForResourceInput) (req *request.Request, output *ListTagsForResourceOutput)
  // ListTagsForResource API operation for Amazon Textract.
  //
  // Lists all tags for an Amazon Textract resource.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation ListTagsForResource for usage and error information.
  //
  // Returned Error Types:
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/ListTagsForResource
func (c *Textract) ListTagsForResource(input *ListTagsForResourceInput) (*ListTagsForResourceOutput, error)
  // ListTagsForResourceWithContext is the same as ListTagsForResource with the addition of
  // the ability to pass a context and additional request options.
  //
  // See ListTagsForResource for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) ListTagsForResourceWithContext(ctx aws.Context, input *ListTagsForResourceInput, opts ...request.Option) (*ListTagsForResourceOutput, error)
  // StartDocumentAnalysisRequest generates a "aws/request.Request" representing the
  // client's request for the StartDocumentAnalysis operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See StartDocumentAnalysis for more information on using the StartDocumentAnalysis
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the StartDocumentAnalysisRequest method.
  //	req, resp := client.StartDocumentAnalysisRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartDocumentAnalysis
func (c *Textract) StartDocumentAnalysisRequest(input *StartDocumentAnalysisInput) (req *request.Request, output *StartDocumentAnalysisOutput)
  // StartDocumentAnalysis API operation for Amazon Textract.
  //
  // Starts the asynchronous analysis of an input document for relationships between
  // detected items such as key-value pairs, tables, and selection elements.
  //
  // StartDocumentAnalysis can analyze text in documents that are in JPEG, PNG,
  // TIFF, and PDF format. The documents are stored in an Amazon S3 bucket. Use
  // DocumentLocation to specify the bucket name and file name of the document.
  //
  // StartDocumentAnalysis returns a job identifier (JobId) that you use to get
  // the results of the operation. When text analysis is finished, Amazon Textract
  // publishes a completion status to the Amazon Simple Notification Service (Amazon
  // SNS) topic that you specify in NotificationChannel. To get the results of
  // the text analysis operation, first check that the status value published
  // to the Amazon SNS topic is SUCCEEDED. If so, call GetDocumentAnalysis, and
  // pass the job identifier (JobId) from the initial call to StartDocumentAnalysis.
  //
  // For more information, see Document Text Analysis (https://docs.aws.amazon.com/textract/latest/dg/how-it-works-analyzing.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation StartDocumentAnalysis for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - IdempotentParameterMismatchException
  //     A ClientRequestToken input parameter was reused with an operation, but at
  //     least one of the other input parameters is different from the previous call
  //     to the operation.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - LimitExceededException
  //     An Amazon Textract service limit was exceeded. For example, if you start
  //     too many asynchronous jobs concurrently, calls to start operations (StartDocumentTextDetection,
  //     for example) raise a LimitExceededException exception (HTTP status code:
  //     400) until the number of concurrently running jobs is below the Amazon Textract
  //     service limit.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartDocumentAnalysis
func (c *Textract) StartDocumentAnalysis(input *StartDocumentAnalysisInput) (*StartDocumentAnalysisOutput, error)
  // StartDocumentAnalysisWithContext is the same as StartDocumentAnalysis with the addition of
  // the ability to pass a context and additional request options.
  //
  // See StartDocumentAnalysis for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) StartDocumentAnalysisWithContext(ctx aws.Context, input *StartDocumentAnalysisInput, opts ...request.Option) (*StartDocumentAnalysisOutput, error)
  // StartDocumentTextDetectionRequest generates a "aws/request.Request" representing the
  // client's request for the StartDocumentTextDetection operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See StartDocumentTextDetection for more information on using the StartDocumentTextDetection
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the StartDocumentTextDetectionRequest method.
  //	req, resp := client.StartDocumentTextDetectionRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartDocumentTextDetection
func (c *Textract) StartDocumentTextDetectionRequest(input *StartDocumentTextDetectionInput) (req *request.Request, output *StartDocumentTextDetectionOutput)
  // StartDocumentTextDetection API operation for Amazon Textract.
  //
  // Starts the asynchronous detection of text in a document. Amazon Textract
  // can detect lines of text and the words that make up a line of text.
  //
  // StartDocumentTextDetection can analyze text in documents that are in JPEG,
  // PNG, TIFF, and PDF format. The documents are stored in an Amazon S3 bucket.
  // Use DocumentLocation to specify the bucket name and file name of the document.
  //
  // StartTextDetection returns a job identifier (JobId) that you use to get the
  // results of the operation. When text detection is finished, Amazon Textract
  // publishes a completion status to the Amazon Simple Notification Service (Amazon
  // SNS) topic that you specify in NotificationChannel. To get the results of
  // the text detection operation, first check that the status value published
  // to the Amazon SNS topic is SUCCEEDED. If so, call GetDocumentTextDetection,
  // and pass the job identifier (JobId) from the initial call to StartDocumentTextDetection.
  //
  // For more information, see Document Text Detection (https://docs.aws.amazon.com/textract/latest/dg/how-it-works-detecting.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation StartDocumentTextDetection for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - IdempotentParameterMismatchException
  //     A ClientRequestToken input parameter was reused with an operation, but at
  //     least one of the other input parameters is different from the previous call
  //     to the operation.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - LimitExceededException
  //     An Amazon Textract service limit was exceeded. For example, if you start
  //     too many asynchronous jobs concurrently, calls to start operations (StartDocumentTextDetection,
  //     for example) raise a LimitExceededException exception (HTTP status code:
  //     400) until the number of concurrently running jobs is below the Amazon Textract
  //     service limit.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartDocumentTextDetection
func (c *Textract) StartDocumentTextDetection(input *StartDocumentTextDetectionInput) (*StartDocumentTextDetectionOutput, error)
  // StartDocumentTextDetectionWithContext is the same as StartDocumentTextDetection with the addition of
  // the ability to pass a context and additional request options.
  //
  // See StartDocumentTextDetection for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) StartDocumentTextDetectionWithContext(ctx aws.Context, input *StartDocumentTextDetectionInput, opts ...request.Option) (*StartDocumentTextDetectionOutput, error)
  // StartExpenseAnalysisRequest generates a "aws/request.Request" representing the
  // client's request for the StartExpenseAnalysis operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See StartExpenseAnalysis for more information on using the StartExpenseAnalysis
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the StartExpenseAnalysisRequest method.
  //	req, resp := client.StartExpenseAnalysisRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartExpenseAnalysis
func (c *Textract) StartExpenseAnalysisRequest(input *StartExpenseAnalysisInput) (req *request.Request, output *StartExpenseAnalysisOutput)
  // StartExpenseAnalysis API operation for Amazon Textract.
  //
  // Starts the asynchronous analysis of invoices or receipts for data like contact
  // information, items purchased, and vendor names.
  //
  // StartExpenseAnalysis can analyze text in documents that are in JPEG, PNG,
  // and PDF format. The documents must be stored in an Amazon S3 bucket. Use
  // the DocumentLocation parameter to specify the name of your S3 bucket and
  // the name of the document in that bucket.
  //
  // StartExpenseAnalysis returns a job identifier (JobId) that you will provide
  // to GetExpenseAnalysis to retrieve the results of the operation. When the
  // analysis of the input invoices/receipts is finished, Amazon Textract publishes
  // a completion status to the Amazon Simple Notification Service (Amazon SNS)
  // topic that you provide to the NotificationChannel. To obtain the results
  // of the invoice and receipt analysis operation, ensure that the status value
  // published to the Amazon SNS topic is SUCCEEDED. If so, call GetExpenseAnalysis,
  // and pass the job identifier (JobId) that was returned by your call to StartExpenseAnalysis.
  //
  // For more information, see Analyzing Invoices and Receipts (https://docs.aws.amazon.com/textract/latest/dg/invoice-receipts.html).
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation StartExpenseAnalysis for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - IdempotentParameterMismatchException
  //     A ClientRequestToken input parameter was reused with an operation, but at
  //     least one of the other input parameters is different from the previous call
  //     to the operation.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - LimitExceededException
  //     An Amazon Textract service limit was exceeded. For example, if you start
  //     too many asynchronous jobs concurrently, calls to start operations (StartDocumentTextDetection,
  //     for example) raise a LimitExceededException exception (HTTP status code:
  //     400) until the number of concurrently running jobs is below the Amazon Textract
  //     service limit.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartExpenseAnalysis
func (c *Textract) StartExpenseAnalysis(input *StartExpenseAnalysisInput) (*StartExpenseAnalysisOutput, error)
  // StartExpenseAnalysisWithContext is the same as StartExpenseAnalysis with the addition of
  // the ability to pass a context and additional request options.
  //
  // See StartExpenseAnalysis for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) StartExpenseAnalysisWithContext(ctx aws.Context, input *StartExpenseAnalysisInput, opts ...request.Option) (*StartExpenseAnalysisOutput, error)
  // StartLendingAnalysisRequest generates a "aws/request.Request" representing the
  // client's request for the StartLendingAnalysis operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See StartLendingAnalysis for more information on using the StartLendingAnalysis
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the StartLendingAnalysisRequest method.
  //	req, resp := client.StartLendingAnalysisRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartLendingAnalysis
func (c *Textract) StartLendingAnalysisRequest(input *StartLendingAnalysisInput) (req *request.Request, output *StartLendingAnalysisOutput)
  // StartLendingAnalysis API operation for Amazon Textract.
  //
  // Starts the classification and analysis of an input document. StartLendingAnalysis
  // initiates the classification and analysis of a packet of lending documents.
  // StartLendingAnalysis operates on a document file located in an Amazon S3
  // bucket.
  //
  // StartLendingAnalysis can analyze text in documents that are in one of the
  // following formats: JPEG, PNG, TIFF, PDF. Use DocumentLocation to specify
  // the bucket name and the file name of the document.
  //
  // StartLendingAnalysis returns a job identifier (JobId) that you use to get
  // the results of the operation. When the text analysis is finished, Amazon
  // Textract publishes a completion status to the Amazon Simple Notification
  // Service (Amazon SNS) topic that you specify in NotificationChannel. To get
  // the results of the text analysis operation, first check that the status value
  // published to the Amazon SNS topic is SUCCEEDED. If the status is SUCCEEDED
  // you can call either GetLendingAnalysis or GetLendingAnalysisSummary and provide
  // the JobId to obtain the results of the analysis.
  //
  // If using OutputConfig to specify an Amazon S3 bucket, the output will be
  // contained within the specified prefix in a directory labeled with the job-id.
  // In the directory there are 3 sub-directories:
  //
  //   - detailedResponse (contains the GetLendingAnalysis response)
  //
  //   - summaryResponse (for the GetLendingAnalysisSummary response)
  //
  //   - splitDocuments (documents split across logical boundaries)
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation StartLendingAnalysis for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - InvalidS3ObjectException
  //     Amazon Textract is unable to access the S3 object that's specified in the
  //     request. for more information, Configure Access to Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html)
  //     For troubleshooting information, see Troubleshooting Amazon S3 (https://docs.aws.amazon.com/AmazonS3/latest/dev/troubleshooting.html)
  //
  //   - InvalidKMSKeyException
  //     Indicates you do not have decrypt permissions with the KMS key entered, or
  //     the KMS key was entered incorrectly.
  //
  //   - UnsupportedDocumentException
  //     The format of the input document isn't supported. Documents for operations
  //     can be in PNG, JPEG, PDF, or TIFF format.
  //
  //   - DocumentTooLargeException
  //     The document can't be processed because it's too large. The maximum document
  //     size for synchronous operations 10 MB. The maximum document size for asynchronous
  //     operations is 500 MB for PDF files.
  //
  //   - BadDocumentException
  //     Amazon Textract isn't able to read the document. For more information on
  //     the document limits in Amazon Textract, see limits.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - IdempotentParameterMismatchException
  //     A ClientRequestToken input parameter was reused with an operation, but at
  //     least one of the other input parameters is different from the previous call
  //     to the operation.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - LimitExceededException
  //     An Amazon Textract service limit was exceeded. For example, if you start
  //     too many asynchronous jobs concurrently, calls to start operations (StartDocumentTextDetection,
  //     for example) raise a LimitExceededException exception (HTTP status code:
  //     400) until the number of concurrently running jobs is below the Amazon Textract
  //     service limit.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/StartLendingAnalysis
func (c *Textract) StartLendingAnalysis(input *StartLendingAnalysisInput) (*StartLendingAnalysisOutput, error)
  // StartLendingAnalysisWithContext is the same as StartLendingAnalysis with the addition of
  // the ability to pass a context and additional request options.
  //
  // See StartLendingAnalysis for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) StartLendingAnalysisWithContext(ctx aws.Context, input *StartLendingAnalysisInput, opts ...request.Option) (*StartLendingAnalysisOutput, error)
  // TagResourceRequest generates a "aws/request.Request" representing the
  // client's request for the TagResource operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See TagResource for more information on using the TagResource
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the TagResourceRequest method.
  //	req, resp := client.TagResourceRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/TagResource
func (c *Textract) TagResourceRequest(input *TagResourceInput) (req *request.Request, output *TagResourceOutput)
  // TagResource API operation for Amazon Textract.
  //
  // Adds one or more tags to the specified resource.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation TagResource for usage and error information.
  //
  // Returned Error Types:
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - ServiceQuotaExceededException
  //     Returned when a request cannot be completed as it would exceed a maximum
  //     service quota.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/TagResource
func (c *Textract) TagResource(input *TagResourceInput) (*TagResourceOutput, error)
  // TagResourceWithContext is the same as TagResource with the addition of
  // the ability to pass a context and additional request options.
  //
  // See TagResource for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) TagResourceWithContext(ctx aws.Context, input *TagResourceInput, opts ...request.Option) (*TagResourceOutput, error)
  // UntagResourceRequest generates a "aws/request.Request" representing the
  // client's request for the UntagResource operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See UntagResource for more information on using the UntagResource
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the UntagResourceRequest method.
  //	req, resp := client.UntagResourceRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/UntagResource
func (c *Textract) UntagResourceRequest(input *UntagResourceInput) (req *request.Request, output *UntagResourceOutput)
  // UntagResource API operation for Amazon Textract.
  //
  // Removes any tags with the specified keys from the specified resource.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation UntagResource for usage and error information.
  //
  // Returned Error Types:
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/UntagResource
func (c *Textract) UntagResource(input *UntagResourceInput) (*UntagResourceOutput, error)
  // UntagResourceWithContext is the same as UntagResource with the addition of
  // the ability to pass a context and additional request options.
  //
  // See UntagResource for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) UntagResourceWithContext(ctx aws.Context, input *UntagResourceInput, opts ...request.Option) (*UntagResourceOutput, error)
  // UpdateAdapterRequest generates a "aws/request.Request" representing the
  // client's request for the UpdateAdapter operation. The "output" return
  // value will be populated with the request's response once the request completes
  // successfully.
  //
  // Use "Send" method on the returned Request to send the API call to the service.
  // the "output" return value is not valid until after Send returns without error.
  //
  // See UpdateAdapter for more information on using the UpdateAdapter
  // API call, and error handling.
  //
  // This method is useful when you want to inject custom logic or configuration
  // into the SDK's request lifecycle. Such as custom headers, or retry logic.
  //
  //	// Example sending a request using the UpdateAdapterRequest method.
  //	req, resp := client.UpdateAdapterRequest(params)
  //
  //	err := req.Send()
  //	if err == nil { // resp is now filled
  //	    fmt.Println(resp)
  //	}
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/UpdateAdapter
func (c *Textract) UpdateAdapterRequest(input *UpdateAdapterInput) (req *request.Request, output *UpdateAdapterOutput)
  // UpdateAdapter API operation for Amazon Textract.
  //
  // Update the configuration for an adapter. FeatureTypes configurations cannot
  // be updated. At least one new parameter must be specified as an argument.
  //
  // Returns awserr.Error for service API and SDK errors. Use runtime type assertions
  // with awserr.Error's Code and Message methods to get detailed information about
  // the error.
  //
  // See the AWS API reference guide for Amazon Textract's
  // API operation UpdateAdapter for usage and error information.
  //
  // Returned Error Types:
  //
  //   - InvalidParameterException
  //     An input parameter violated a constraint. For example, in synchronous operations,
  //     an InvalidParameterException exception occurs when neither of the S3Object
  //     or Bytes values are supplied in the Document request parameter. Validate
  //     your parameter before calling the API operation again.
  //
  //   - AccessDeniedException
  //     You aren't authorized to perform the action. Use the Amazon Resource Name
  //     (ARN) of an authorized user or IAM role to perform the operation.
  //
  //   - ConflictException
  //     Updating or deleting a resource can cause an inconsistent state.
  //
  //   - ProvisionedThroughputExceededException
  //     The number of requests exceeded your throughput limit. If you want to increase
  //     this limit, contact Amazon Textract.
  //
  //   - InternalServerError
  //     Amazon Textract experienced a service issue. Try your call again.
  //
  //   - ThrottlingException
  //     Amazon Textract is temporarily unable to process the request. Try your call
  //     again.
  //
  //   - ValidationException
  //     Indicates that a request was not valid. Check request for proper formatting.
  //
  //   - ResourceNotFoundException
  //     Returned when an operation tried to access a nonexistent resource.
  //
  // See also, https://docs.aws.amazon.com/goto/WebAPI/textract-2018-06-27/UpdateAdapter
func (c *Textract) UpdateAdapter(input *UpdateAdapterInput) (*UpdateAdapterOutput, error)
  // UpdateAdapterWithContext is the same as UpdateAdapter with the addition of
  // the ability to pass a context and additional request options.
  //
  // See UpdateAdapter for details on how to use this API operation.
  //
  // The context must be non-nil and will be used for request cancellation. If
  // the context is nil a panic will occur. In the future the SDK may create
  // sub-contexts for http.Requests. See https://golang.org/pkg/context/
  // for more information on using Contexts.
func (c *Textract) UpdateAdapterWithContext(ctx aws.Context, input *UpdateAdapterInput, opts ...request.Option) (*UpdateAdapterOutput, error)
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AccessDeniedException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AccessDeniedException) GoString() string
  // Code returns the exception type name.
func (s *AccessDeniedException) Code() string
  // Message returns the exception's message.
func (s *AccessDeniedException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *AccessDeniedException) OrigErr() error
func (s *AccessDeniedException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *AccessDeniedException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *AccessDeniedException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Adapter) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Adapter) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *Adapter) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *Adapter) SetAdapterId(v string) *Adapter
  // SetPages sets the Pages field's value.
func (s *Adapter) SetPages(v []*string) *Adapter
  // SetVersion sets the Version field's value.
func (s *Adapter) SetVersion(v string) *Adapter
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterOverview) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterOverview) GoString() string
  // SetAdapterId sets the AdapterId field's value.
func (s *AdapterOverview) SetAdapterId(v string) *AdapterOverview
  // SetAdapterName sets the AdapterName field's value.
func (s *AdapterOverview) SetAdapterName(v string) *AdapterOverview
  // SetCreationTime sets the CreationTime field's value.
func (s *AdapterOverview) SetCreationTime(v time.Time) *AdapterOverview
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *AdapterOverview) SetFeatureTypes(v []*string) *AdapterOverview
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterVersionDatasetConfig) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterVersionDatasetConfig) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *AdapterVersionDatasetConfig) Validate() error
  // SetManifestS3Object sets the ManifestS3Object field's value.
func (s *AdapterVersionDatasetConfig) SetManifestS3Object(v *S3Object) *AdapterVersionDatasetConfig
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterVersionEvaluationMetric) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterVersionEvaluationMetric) GoString() string
  // SetAdapterVersion sets the AdapterVersion field's value.
func (s *AdapterVersionEvaluationMetric) SetAdapterVersion(v *EvaluationMetric) *AdapterVersionEvaluationMetric
  // SetBaseline sets the Baseline field's value.
func (s *AdapterVersionEvaluationMetric) SetBaseline(v *EvaluationMetric) *AdapterVersionEvaluationMetric
  // SetFeatureType sets the FeatureType field's value.
func (s *AdapterVersionEvaluationMetric) SetFeatureType(v string) *AdapterVersionEvaluationMetric
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterVersionOverview) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdapterVersionOverview) GoString() string
  // SetAdapterId sets the AdapterId field's value.
func (s *AdapterVersionOverview) SetAdapterId(v string) *AdapterVersionOverview
  // SetAdapterVersion sets the AdapterVersion field's value.
func (s *AdapterVersionOverview) SetAdapterVersion(v string) *AdapterVersionOverview
  // SetCreationTime sets the CreationTime field's value.
func (s *AdapterVersionOverview) SetCreationTime(v time.Time) *AdapterVersionOverview
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *AdapterVersionOverview) SetFeatureTypes(v []*string) *AdapterVersionOverview
  // SetStatus sets the Status field's value.
func (s *AdapterVersionOverview) SetStatus(v string) *AdapterVersionOverview
  // SetStatusMessage sets the StatusMessage field's value.
func (s *AdapterVersionOverview) SetStatusMessage(v string) *AdapterVersionOverview
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdaptersConfig) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AdaptersConfig) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *AdaptersConfig) Validate() error
  // SetAdapters sets the Adapters field's value.
func (s *AdaptersConfig) SetAdapters(v []*Adapter) *AdaptersConfig
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeDocumentInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeDocumentInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *AnalyzeDocumentInput) Validate() error
  // SetAdaptersConfig sets the AdaptersConfig field's value.
func (s *AnalyzeDocumentInput) SetAdaptersConfig(v *AdaptersConfig) *AnalyzeDocumentInput
  // SetDocument sets the Document field's value.
func (s *AnalyzeDocumentInput) SetDocument(v *Document) *AnalyzeDocumentInput
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *AnalyzeDocumentInput) SetFeatureTypes(v []*string) *AnalyzeDocumentInput
  // SetHumanLoopConfig sets the HumanLoopConfig field's value.
func (s *AnalyzeDocumentInput) SetHumanLoopConfig(v *HumanLoopConfig) *AnalyzeDocumentInput
  // SetQueriesConfig sets the QueriesConfig field's value.
func (s *AnalyzeDocumentInput) SetQueriesConfig(v *QueriesConfig) *AnalyzeDocumentInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeDocumentOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeDocumentOutput) GoString() string
  // SetAnalyzeDocumentModelVersion sets the AnalyzeDocumentModelVersion field's value.
func (s *AnalyzeDocumentOutput) SetAnalyzeDocumentModelVersion(v string) *AnalyzeDocumentOutput
  // SetBlocks sets the Blocks field's value.
func (s *AnalyzeDocumentOutput) SetBlocks(v []*Block) *AnalyzeDocumentOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *AnalyzeDocumentOutput) SetDocumentMetadata(v *DocumentMetadata) *AnalyzeDocumentOutput
  // SetHumanLoopActivationOutput sets the HumanLoopActivationOutput field's value.
func (s *AnalyzeDocumentOutput) SetHumanLoopActivationOutput(v *HumanLoopActivationOutput) *AnalyzeDocumentOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeExpenseInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeExpenseInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *AnalyzeExpenseInput) Validate() error
  // SetDocument sets the Document field's value.
func (s *AnalyzeExpenseInput) SetDocument(v *Document) *AnalyzeExpenseInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeExpenseOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeExpenseOutput) GoString() string
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *AnalyzeExpenseOutput) SetDocumentMetadata(v *DocumentMetadata) *AnalyzeExpenseOutput
  // SetExpenseDocuments sets the ExpenseDocuments field's value.
func (s *AnalyzeExpenseOutput) SetExpenseDocuments(v []*ExpenseDocument) *AnalyzeExpenseOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeIDDetections) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeIDDetections) GoString() string
  // SetConfidence sets the Confidence field's value.
func (s *AnalyzeIDDetections) SetConfidence(v float64) *AnalyzeIDDetections
  // SetNormalizedValue sets the NormalizedValue field's value.
func (s *AnalyzeIDDetections) SetNormalizedValue(v *NormalizedValue) *AnalyzeIDDetections
  // SetText sets the Text field's value.
func (s *AnalyzeIDDetections) SetText(v string) *AnalyzeIDDetections
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeIDInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeIDInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *AnalyzeIDInput) Validate() error
  // SetDocumentPages sets the DocumentPages field's value.
func (s *AnalyzeIDInput) SetDocumentPages(v []*Document) *AnalyzeIDInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeIDOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s AnalyzeIDOutput) GoString() string
  // SetAnalyzeIDModelVersion sets the AnalyzeIDModelVersion field's value.
func (s *AnalyzeIDOutput) SetAnalyzeIDModelVersion(v string) *AnalyzeIDOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *AnalyzeIDOutput) SetDocumentMetadata(v *DocumentMetadata) *AnalyzeIDOutput
  // SetIdentityDocuments sets the IdentityDocuments field's value.
func (s *AnalyzeIDOutput) SetIdentityDocuments(v []*IdentityDocument) *AnalyzeIDOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s BadDocumentException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s BadDocumentException) GoString() string
  // Code returns the exception type name.
func (s *BadDocumentException) Code() string
  // Message returns the exception's message.
func (s *BadDocumentException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *BadDocumentException) OrigErr() error
func (s *BadDocumentException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *BadDocumentException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *BadDocumentException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Block) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Block) GoString() string
  // SetBlockType sets the BlockType field's value.
func (s *Block) SetBlockType(v string) *Block
  // SetColumnIndex sets the ColumnIndex field's value.
func (s *Block) SetColumnIndex(v int64) *Block
  // SetColumnSpan sets the ColumnSpan field's value.
func (s *Block) SetColumnSpan(v int64) *Block
  // SetConfidence sets the Confidence field's value.
func (s *Block) SetConfidence(v float64) *Block
  // SetEntityTypes sets the EntityTypes field's value.
func (s *Block) SetEntityTypes(v []*string) *Block
  // SetGeometry sets the Geometry field's value.
func (s *Block) SetGeometry(v *Geometry) *Block
  // SetId sets the Id field's value.
func (s *Block) SetId(v string) *Block
  // SetPage sets the Page field's value.
func (s *Block) SetPage(v int64) *Block
  // SetQuery sets the Query field's value.
func (s *Block) SetQuery(v *Query) *Block
  // SetRelationships sets the Relationships field's value.
func (s *Block) SetRelationships(v []*Relationship) *Block
  // SetRowIndex sets the RowIndex field's value.
func (s *Block) SetRowIndex(v int64) *Block
  // SetRowSpan sets the RowSpan field's value.
func (s *Block) SetRowSpan(v int64) *Block
  // SetSelectionStatus sets the SelectionStatus field's value.
func (s *Block) SetSelectionStatus(v string) *Block
  // SetText sets the Text field's value.
func (s *Block) SetText(v string) *Block
  // SetTextType sets the TextType field's value.
func (s *Block) SetTextType(v string) *Block
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s BoundingBox) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s BoundingBox) GoString() string
  // SetHeight sets the Height field's value.
func (s *BoundingBox) SetHeight(v float64) *BoundingBox
  // SetLeft sets the Left field's value.
func (s *BoundingBox) SetLeft(v float64) *BoundingBox
  // SetTop sets the Top field's value.
func (s *BoundingBox) SetTop(v float64) *BoundingBox
  // SetWidth sets the Width field's value.
func (s *BoundingBox) SetWidth(v float64) *BoundingBox
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ConflictException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ConflictException) GoString() string
  // Code returns the exception type name.
func (s *ConflictException) Code() string
  // Message returns the exception's message.
func (s *ConflictException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ConflictException) OrigErr() error
func (s *ConflictException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *ConflictException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *ConflictException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *CreateAdapterInput) Validate() error
  // SetAdapterName sets the AdapterName field's value.
func (s *CreateAdapterInput) SetAdapterName(v string) *CreateAdapterInput
  // SetAutoUpdate sets the AutoUpdate field's value.
func (s *CreateAdapterInput) SetAutoUpdate(v string) *CreateAdapterInput
  // SetClientRequestToken sets the ClientRequestToken field's value.
func (s *CreateAdapterInput) SetClientRequestToken(v string) *CreateAdapterInput
  // SetDescription sets the Description field's value.
func (s *CreateAdapterInput) SetDescription(v string) *CreateAdapterInput
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *CreateAdapterInput) SetFeatureTypes(v []*string) *CreateAdapterInput
  // SetTags sets the Tags field's value.
func (s *CreateAdapterInput) SetTags(v map[string]*string) *CreateAdapterInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterOutput) GoString() string
  // SetAdapterId sets the AdapterId field's value.
func (s *CreateAdapterOutput) SetAdapterId(v string) *CreateAdapterOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterVersionInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterVersionInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *CreateAdapterVersionInput) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *CreateAdapterVersionInput) SetAdapterId(v string) *CreateAdapterVersionInput
  // SetClientRequestToken sets the ClientRequestToken field's value.
func (s *CreateAdapterVersionInput) SetClientRequestToken(v string) *CreateAdapterVersionInput
  // SetDatasetConfig sets the DatasetConfig field's value.
func (s *CreateAdapterVersionInput) SetDatasetConfig(v *AdapterVersionDatasetConfig) *CreateAdapterVersionInput
  // SetKMSKeyId sets the KMSKeyId field's value.
func (s *CreateAdapterVersionInput) SetKMSKeyId(v string) *CreateAdapterVersionInput
  // SetOutputConfig sets the OutputConfig field's value.
func (s *CreateAdapterVersionInput) SetOutputConfig(v *OutputConfig) *CreateAdapterVersionInput
  // SetTags sets the Tags field's value.
func (s *CreateAdapterVersionInput) SetTags(v map[string]*string) *CreateAdapterVersionInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterVersionOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s CreateAdapterVersionOutput) GoString() string
  // SetAdapterId sets the AdapterId field's value.
func (s *CreateAdapterVersionOutput) SetAdapterId(v string) *CreateAdapterVersionOutput
  // SetAdapterVersion sets the AdapterVersion field's value.
func (s *CreateAdapterVersionOutput) SetAdapterVersion(v string) *CreateAdapterVersionOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *DeleteAdapterInput) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *DeleteAdapterInput) SetAdapterId(v string) *DeleteAdapterInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterOutput) GoString() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterVersionInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterVersionInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *DeleteAdapterVersionInput) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *DeleteAdapterVersionInput) SetAdapterId(v string) *DeleteAdapterVersionInput
  // SetAdapterVersion sets the AdapterVersion field's value.
func (s *DeleteAdapterVersionInput) SetAdapterVersion(v string) *DeleteAdapterVersionInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterVersionOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DeleteAdapterVersionOutput) GoString() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DetectDocumentTextInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DetectDocumentTextInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *DetectDocumentTextInput) Validate() error
  // SetDocument sets the Document field's value.
func (s *DetectDocumentTextInput) SetDocument(v *Document) *DetectDocumentTextInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DetectDocumentTextOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DetectDocumentTextOutput) GoString() string
  // SetBlocks sets the Blocks field's value.
func (s *DetectDocumentTextOutput) SetBlocks(v []*Block) *DetectDocumentTextOutput
  // SetDetectDocumentTextModelVersion sets the DetectDocumentTextModelVersion field's value.
func (s *DetectDocumentTextOutput) SetDetectDocumentTextModelVersion(v string) *DetectDocumentTextOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *DetectDocumentTextOutput) SetDocumentMetadata(v *DocumentMetadata) *DetectDocumentTextOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DetectedSignature) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DetectedSignature) GoString() string
  // SetPage sets the Page field's value.
func (s *DetectedSignature) SetPage(v int64) *DetectedSignature
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Document) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Document) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *Document) Validate() error
  // SetBytes sets the Bytes field's value.
func (s *Document) SetBytes(v []byte) *Document
  // SetS3Object sets the S3Object field's value.
func (s *Document) SetS3Object(v *S3Object) *Document
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentGroup) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentGroup) GoString() string
  // SetDetectedSignatures sets the DetectedSignatures field's value.
func (s *DocumentGroup) SetDetectedSignatures(v []*DetectedSignature) *DocumentGroup
  // SetSplitDocuments sets the SplitDocuments field's value.
func (s *DocumentGroup) SetSplitDocuments(v []*SplitDocument) *DocumentGroup
  // SetType sets the Type field's value.
func (s *DocumentGroup) SetType(v string) *DocumentGroup
  // SetUndetectedSignatures sets the UndetectedSignatures field's value.
func (s *DocumentGroup) SetUndetectedSignatures(v []*UndetectedSignature) *DocumentGroup
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentLocation) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentLocation) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *DocumentLocation) Validate() error
  // SetS3Object sets the S3Object field's value.
func (s *DocumentLocation) SetS3Object(v *S3Object) *DocumentLocation
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentMetadata) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentMetadata) GoString() string
  // SetPages sets the Pages field's value.
func (s *DocumentMetadata) SetPages(v int64) *DocumentMetadata
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentTooLargeException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s DocumentTooLargeException) GoString() string
  // Code returns the exception type name.
func (s *DocumentTooLargeException) Code() string
  // Message returns the exception's message.
func (s *DocumentTooLargeException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *DocumentTooLargeException) OrigErr() error
func (s *DocumentTooLargeException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *DocumentTooLargeException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *DocumentTooLargeException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s EvaluationMetric) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s EvaluationMetric) GoString() string
  // SetF1Score sets the F1Score field's value.
func (s *EvaluationMetric) SetF1Score(v float64) *EvaluationMetric
  // SetPrecision sets the Precision field's value.
func (s *EvaluationMetric) SetPrecision(v float64) *EvaluationMetric
  // SetRecall sets the Recall field's value.
func (s *EvaluationMetric) SetRecall(v float64) *EvaluationMetric
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseCurrency) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseCurrency) GoString() string
  // SetCode sets the Code field's value.
func (s *ExpenseCurrency) SetCode(v string) *ExpenseCurrency
  // SetConfidence sets the Confidence field's value.
func (s *ExpenseCurrency) SetConfidence(v float64) *ExpenseCurrency
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseDetection) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseDetection) GoString() string
  // SetConfidence sets the Confidence field's value.
func (s *ExpenseDetection) SetConfidence(v float64) *ExpenseDetection
  // SetGeometry sets the Geometry field's value.
func (s *ExpenseDetection) SetGeometry(v *Geometry) *ExpenseDetection
  // SetText sets the Text field's value.
func (s *ExpenseDetection) SetText(v string) *ExpenseDetection
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseDocument) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseDocument) GoString() string
  // SetBlocks sets the Blocks field's value.
func (s *ExpenseDocument) SetBlocks(v []*Block) *ExpenseDocument
  // SetExpenseIndex sets the ExpenseIndex field's value.
func (s *ExpenseDocument) SetExpenseIndex(v int64) *ExpenseDocument
  // SetLineItemGroups sets the LineItemGroups field's value.
func (s *ExpenseDocument) SetLineItemGroups(v []*LineItemGroup) *ExpenseDocument
  // SetSummaryFields sets the SummaryFields field's value.
func (s *ExpenseDocument) SetSummaryFields(v []*ExpenseField) *ExpenseDocument
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseField) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseField) GoString() string
  // SetCurrency sets the Currency field's value.
func (s *ExpenseField) SetCurrency(v *ExpenseCurrency) *ExpenseField
  // SetGroupProperties sets the GroupProperties field's value.
func (s *ExpenseField) SetGroupProperties(v []*ExpenseGroupProperty) *ExpenseField
  // SetLabelDetection sets the LabelDetection field's value.
func (s *ExpenseField) SetLabelDetection(v *ExpenseDetection) *ExpenseField
  // SetPageNumber sets the PageNumber field's value.
func (s *ExpenseField) SetPageNumber(v int64) *ExpenseField
  // SetType sets the Type field's value.
func (s *ExpenseField) SetType(v *ExpenseType) *ExpenseField
  // SetValueDetection sets the ValueDetection field's value.
func (s *ExpenseField) SetValueDetection(v *ExpenseDetection) *ExpenseField
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseGroupProperty) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseGroupProperty) GoString() string
  // SetId sets the Id field's value.
func (s *ExpenseGroupProperty) SetId(v string) *ExpenseGroupProperty
  // SetTypes sets the Types field's value.
func (s *ExpenseGroupProperty) SetTypes(v []*string) *ExpenseGroupProperty
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseType) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ExpenseType) GoString() string
  // SetConfidence sets the Confidence field's value.
func (s *ExpenseType) SetConfidence(v float64) *ExpenseType
  // SetText sets the Text field's value.
func (s *ExpenseType) SetText(v string) *ExpenseType
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Extraction) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Extraction) GoString() string
  // SetExpenseDocument sets the ExpenseDocument field's value.
func (s *Extraction) SetExpenseDocument(v *ExpenseDocument) *Extraction
  // SetIdentityDocument sets the IdentityDocument field's value.
func (s *Extraction) SetIdentityDocument(v *IdentityDocument) *Extraction
  // SetLendingDocument sets the LendingDocument field's value.
func (s *Extraction) SetLendingDocument(v *LendingDocument) *Extraction
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Geometry) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Geometry) GoString() string
  // SetBoundingBox sets the BoundingBox field's value.
func (s *Geometry) SetBoundingBox(v *BoundingBox) *Geometry
  // SetPolygon sets the Polygon field's value.
func (s *Geometry) SetPolygon(v []*Point) *Geometry
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *GetAdapterInput) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *GetAdapterInput) SetAdapterId(v string) *GetAdapterInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterOutput) GoString() string
  // SetAdapterId sets the AdapterId field's value.
func (s *GetAdapterOutput) SetAdapterId(v string) *GetAdapterOutput
  // SetAdapterName sets the AdapterName field's value.
func (s *GetAdapterOutput) SetAdapterName(v string) *GetAdapterOutput
  // SetAutoUpdate sets the AutoUpdate field's value.
func (s *GetAdapterOutput) SetAutoUpdate(v string) *GetAdapterOutput
  // SetCreationTime sets the CreationTime field's value.
func (s *GetAdapterOutput) SetCreationTime(v time.Time) *GetAdapterOutput
  // SetDescription sets the Description field's value.
func (s *GetAdapterOutput) SetDescription(v string) *GetAdapterOutput
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *GetAdapterOutput) SetFeatureTypes(v []*string) *GetAdapterOutput
  // SetTags sets the Tags field's value.
func (s *GetAdapterOutput) SetTags(v map[string]*string) *GetAdapterOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterVersionInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterVersionInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *GetAdapterVersionInput) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *GetAdapterVersionInput) SetAdapterId(v string) *GetAdapterVersionInput
  // SetAdapterVersion sets the AdapterVersion field's value.
func (s *GetAdapterVersionInput) SetAdapterVersion(v string) *GetAdapterVersionInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterVersionOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetAdapterVersionOutput) GoString() string
  // SetAdapterId sets the AdapterId field's value.
func (s *GetAdapterVersionOutput) SetAdapterId(v string) *GetAdapterVersionOutput
  // SetAdapterVersion sets the AdapterVersion field's value.
func (s *GetAdapterVersionOutput) SetAdapterVersion(v string) *GetAdapterVersionOutput
  // SetCreationTime sets the CreationTime field's value.
func (s *GetAdapterVersionOutput) SetCreationTime(v time.Time) *GetAdapterVersionOutput
  // SetDatasetConfig sets the DatasetConfig field's value.
func (s *GetAdapterVersionOutput) SetDatasetConfig(v *AdapterVersionDatasetConfig) *GetAdapterVersionOutput
  // SetEvaluationMetrics sets the EvaluationMetrics field's value.
func (s *GetAdapterVersionOutput) SetEvaluationMetrics(v []*AdapterVersionEvaluationMetric) *GetAdapterVersionOutput
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *GetAdapterVersionOutput) SetFeatureTypes(v []*string) *GetAdapterVersionOutput
  // SetKMSKeyId sets the KMSKeyId field's value.
func (s *GetAdapterVersionOutput) SetKMSKeyId(v string) *GetAdapterVersionOutput
  // SetOutputConfig sets the OutputConfig field's value.
func (s *GetAdapterVersionOutput) SetOutputConfig(v *OutputConfig) *GetAdapterVersionOutput
  // SetStatus sets the Status field's value.
func (s *GetAdapterVersionOutput) SetStatus(v string) *GetAdapterVersionOutput
  // SetStatusMessage sets the StatusMessage field's value.
func (s *GetAdapterVersionOutput) SetStatusMessage(v string) *GetAdapterVersionOutput
  // SetTags sets the Tags field's value.
func (s *GetAdapterVersionOutput) SetTags(v map[string]*string) *GetAdapterVersionOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentAnalysisInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentAnalysisInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *GetDocumentAnalysisInput) Validate() error
  // SetJobId sets the JobId field's value.
func (s *GetDocumentAnalysisInput) SetJobId(v string) *GetDocumentAnalysisInput
  // SetMaxResults sets the MaxResults field's value.
func (s *GetDocumentAnalysisInput) SetMaxResults(v int64) *GetDocumentAnalysisInput
  // SetNextToken sets the NextToken field's value.
func (s *GetDocumentAnalysisInput) SetNextToken(v string) *GetDocumentAnalysisInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentAnalysisOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentAnalysisOutput) GoString() string
  // SetAnalyzeDocumentModelVersion sets the AnalyzeDocumentModelVersion field's value.
func (s *GetDocumentAnalysisOutput) SetAnalyzeDocumentModelVersion(v string) *GetDocumentAnalysisOutput
  // SetBlocks sets the Blocks field's value.
func (s *GetDocumentAnalysisOutput) SetBlocks(v []*Block) *GetDocumentAnalysisOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *GetDocumentAnalysisOutput) SetDocumentMetadata(v *DocumentMetadata) *GetDocumentAnalysisOutput
  // SetJobStatus sets the JobStatus field's value.
func (s *GetDocumentAnalysisOutput) SetJobStatus(v string) *GetDocumentAnalysisOutput
  // SetNextToken sets the NextToken field's value.
func (s *GetDocumentAnalysisOutput) SetNextToken(v string) *GetDocumentAnalysisOutput
  // SetStatusMessage sets the StatusMessage field's value.
func (s *GetDocumentAnalysisOutput) SetStatusMessage(v string) *GetDocumentAnalysisOutput
  // SetWarnings sets the Warnings field's value.
func (s *GetDocumentAnalysisOutput) SetWarnings(v []*Warning) *GetDocumentAnalysisOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentTextDetectionInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentTextDetectionInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *GetDocumentTextDetectionInput) Validate() error
  // SetJobId sets the JobId field's value.
func (s *GetDocumentTextDetectionInput) SetJobId(v string) *GetDocumentTextDetectionInput
  // SetMaxResults sets the MaxResults field's value.
func (s *GetDocumentTextDetectionInput) SetMaxResults(v int64) *GetDocumentTextDetectionInput
  // SetNextToken sets the NextToken field's value.
func (s *GetDocumentTextDetectionInput) SetNextToken(v string) *GetDocumentTextDetectionInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentTextDetectionOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetDocumentTextDetectionOutput) GoString() string
  // SetBlocks sets the Blocks field's value.
func (s *GetDocumentTextDetectionOutput) SetBlocks(v []*Block) *GetDocumentTextDetectionOutput
  // SetDetectDocumentTextModelVersion sets the DetectDocumentTextModelVersion field's value.
func (s *GetDocumentTextDetectionOutput) SetDetectDocumentTextModelVersion(v string) *GetDocumentTextDetectionOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *GetDocumentTextDetectionOutput) SetDocumentMetadata(v *DocumentMetadata) *GetDocumentTextDetectionOutput
  // SetJobStatus sets the JobStatus field's value.
func (s *GetDocumentTextDetectionOutput) SetJobStatus(v string) *GetDocumentTextDetectionOutput
  // SetNextToken sets the NextToken field's value.
func (s *GetDocumentTextDetectionOutput) SetNextToken(v string) *GetDocumentTextDetectionOutput
  // SetStatusMessage sets the StatusMessage field's value.
func (s *GetDocumentTextDetectionOutput) SetStatusMessage(v string) *GetDocumentTextDetectionOutput
  // SetWarnings sets the Warnings field's value.
func (s *GetDocumentTextDetectionOutput) SetWarnings(v []*Warning) *GetDocumentTextDetectionOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetExpenseAnalysisInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetExpenseAnalysisInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *GetExpenseAnalysisInput) Validate() error
  // SetJobId sets the JobId field's value.
func (s *GetExpenseAnalysisInput) SetJobId(v string) *GetExpenseAnalysisInput
  // SetMaxResults sets the MaxResults field's value.
func (s *GetExpenseAnalysisInput) SetMaxResults(v int64) *GetExpenseAnalysisInput
  // SetNextToken sets the NextToken field's value.
func (s *GetExpenseAnalysisInput) SetNextToken(v string) *GetExpenseAnalysisInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetExpenseAnalysisOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetExpenseAnalysisOutput) GoString() string
  // SetAnalyzeExpenseModelVersion sets the AnalyzeExpenseModelVersion field's value.
func (s *GetExpenseAnalysisOutput) SetAnalyzeExpenseModelVersion(v string) *GetExpenseAnalysisOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *GetExpenseAnalysisOutput) SetDocumentMetadata(v *DocumentMetadata) *GetExpenseAnalysisOutput
  // SetExpenseDocuments sets the ExpenseDocuments field's value.
func (s *GetExpenseAnalysisOutput) SetExpenseDocuments(v []*ExpenseDocument) *GetExpenseAnalysisOutput
  // SetJobStatus sets the JobStatus field's value.
func (s *GetExpenseAnalysisOutput) SetJobStatus(v string) *GetExpenseAnalysisOutput
  // SetNextToken sets the NextToken field's value.
func (s *GetExpenseAnalysisOutput) SetNextToken(v string) *GetExpenseAnalysisOutput
  // SetStatusMessage sets the StatusMessage field's value.
func (s *GetExpenseAnalysisOutput) SetStatusMessage(v string) *GetExpenseAnalysisOutput
  // SetWarnings sets the Warnings field's value.
func (s *GetExpenseAnalysisOutput) SetWarnings(v []*Warning) *GetExpenseAnalysisOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *GetLendingAnalysisInput) Validate() error
  // SetJobId sets the JobId field's value.
func (s *GetLendingAnalysisInput) SetJobId(v string) *GetLendingAnalysisInput
  // SetMaxResults sets the MaxResults field's value.
func (s *GetLendingAnalysisInput) SetMaxResults(v int64) *GetLendingAnalysisInput
  // SetNextToken sets the NextToken field's value.
func (s *GetLendingAnalysisInput) SetNextToken(v string) *GetLendingAnalysisInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisOutput) GoString() string
  // SetAnalyzeLendingModelVersion sets the AnalyzeLendingModelVersion field's value.
func (s *GetLendingAnalysisOutput) SetAnalyzeLendingModelVersion(v string) *GetLendingAnalysisOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *GetLendingAnalysisOutput) SetDocumentMetadata(v *DocumentMetadata) *GetLendingAnalysisOutput
  // SetJobStatus sets the JobStatus field's value.
func (s *GetLendingAnalysisOutput) SetJobStatus(v string) *GetLendingAnalysisOutput
  // SetNextToken sets the NextToken field's value.
func (s *GetLendingAnalysisOutput) SetNextToken(v string) *GetLendingAnalysisOutput
  // SetResults sets the Results field's value.
func (s *GetLendingAnalysisOutput) SetResults(v []*LendingResult) *GetLendingAnalysisOutput
  // SetStatusMessage sets the StatusMessage field's value.
func (s *GetLendingAnalysisOutput) SetStatusMessage(v string) *GetLendingAnalysisOutput
  // SetWarnings sets the Warnings field's value.
func (s *GetLendingAnalysisOutput) SetWarnings(v []*Warning) *GetLendingAnalysisOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisSummaryInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisSummaryInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *GetLendingAnalysisSummaryInput) Validate() error
  // SetJobId sets the JobId field's value.
func (s *GetLendingAnalysisSummaryInput) SetJobId(v string) *GetLendingAnalysisSummaryInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisSummaryOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s GetLendingAnalysisSummaryOutput) GoString() string
  // SetAnalyzeLendingModelVersion sets the AnalyzeLendingModelVersion field's value.
func (s *GetLendingAnalysisSummaryOutput) SetAnalyzeLendingModelVersion(v string) *GetLendingAnalysisSummaryOutput
  // SetDocumentMetadata sets the DocumentMetadata field's value.
func (s *GetLendingAnalysisSummaryOutput) SetDocumentMetadata(v *DocumentMetadata) *GetLendingAnalysisSummaryOutput
  // SetJobStatus sets the JobStatus field's value.
func (s *GetLendingAnalysisSummaryOutput) SetJobStatus(v string) *GetLendingAnalysisSummaryOutput
  // SetStatusMessage sets the StatusMessage field's value.
func (s *GetLendingAnalysisSummaryOutput) SetStatusMessage(v string) *GetLendingAnalysisSummaryOutput
  // SetSummary sets the Summary field's value.
func (s *GetLendingAnalysisSummaryOutput) SetSummary(v *LendingSummary) *GetLendingAnalysisSummaryOutput
  // SetWarnings sets the Warnings field's value.
func (s *GetLendingAnalysisSummaryOutput) SetWarnings(v []*Warning) *GetLendingAnalysisSummaryOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopActivationOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopActivationOutput) GoString() string
  // SetHumanLoopActivationConditionsEvaluationResults sets the HumanLoopActivationConditionsEvaluationResults field's value.
func (s *HumanLoopActivationOutput) SetHumanLoopActivationConditionsEvaluationResults(v aws.JSONValue) *HumanLoopActivationOutput
  // SetHumanLoopActivationReasons sets the HumanLoopActivationReasons field's value.
func (s *HumanLoopActivationOutput) SetHumanLoopActivationReasons(v []*string) *HumanLoopActivationOutput
  // SetHumanLoopArn sets the HumanLoopArn field's value.
func (s *HumanLoopActivationOutput) SetHumanLoopArn(v string) *HumanLoopActivationOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopConfig) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopConfig) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *HumanLoopConfig) Validate() error
  // SetDataAttributes sets the DataAttributes field's value.
func (s *HumanLoopConfig) SetDataAttributes(v *HumanLoopDataAttributes) *HumanLoopConfig
  // SetFlowDefinitionArn sets the FlowDefinitionArn field's value.
func (s *HumanLoopConfig) SetFlowDefinitionArn(v string) *HumanLoopConfig
  // SetHumanLoopName sets the HumanLoopName field's value.
func (s *HumanLoopConfig) SetHumanLoopName(v string) *HumanLoopConfig
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopDataAttributes) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopDataAttributes) GoString() string
  // SetContentClassifiers sets the ContentClassifiers field's value.
func (s *HumanLoopDataAttributes) SetContentClassifiers(v []*string) *HumanLoopDataAttributes
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopQuotaExceededException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s HumanLoopQuotaExceededException) GoString() string
  // Code returns the exception type name.
func (s *HumanLoopQuotaExceededException) Code() string
  // Message returns the exception's message.
func (s *HumanLoopQuotaExceededException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *HumanLoopQuotaExceededException) OrigErr() error
func (s *HumanLoopQuotaExceededException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *HumanLoopQuotaExceededException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *HumanLoopQuotaExceededException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s IdempotentParameterMismatchException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s IdempotentParameterMismatchException) GoString() string
  // Code returns the exception type name.
func (s *IdempotentParameterMismatchException) Code() string
  // Message returns the exception's message.
func (s *IdempotentParameterMismatchException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *IdempotentParameterMismatchException) OrigErr() error
func (s *IdempotentParameterMismatchException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *IdempotentParameterMismatchException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *IdempotentParameterMismatchException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s IdentityDocument) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s IdentityDocument) GoString() string
  // SetBlocks sets the Blocks field's value.
func (s *IdentityDocument) SetBlocks(v []*Block) *IdentityDocument
  // SetDocumentIndex sets the DocumentIndex field's value.
func (s *IdentityDocument) SetDocumentIndex(v int64) *IdentityDocument
  // SetIdentityDocumentFields sets the IdentityDocumentFields field's value.
func (s *IdentityDocument) SetIdentityDocumentFields(v []*IdentityDocumentField) *IdentityDocument
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s IdentityDocumentField) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s IdentityDocumentField) GoString() string
  // SetType sets the Type field's value.
func (s *IdentityDocumentField) SetType(v *AnalyzeIDDetections) *IdentityDocumentField
  // SetValueDetection sets the ValueDetection field's value.
func (s *IdentityDocumentField) SetValueDetection(v *AnalyzeIDDetections) *IdentityDocumentField
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InternalServerError) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InternalServerError) GoString() string
  // Code returns the exception type name.
func (s *InternalServerError) Code() string
  // Message returns the exception's message.
func (s *InternalServerError) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *InternalServerError) OrigErr() error
func (s *InternalServerError) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *InternalServerError) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *InternalServerError) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidJobIdException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidJobIdException) GoString() string
  // Code returns the exception type name.
func (s *InvalidJobIdException) Code() string
  // Message returns the exception's message.
func (s *InvalidJobIdException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *InvalidJobIdException) OrigErr() error
func (s *InvalidJobIdException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *InvalidJobIdException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *InvalidJobIdException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidKMSKeyException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidKMSKeyException) GoString() string
  // Code returns the exception type name.
func (s *InvalidKMSKeyException) Code() string
  // Message returns the exception's message.
func (s *InvalidKMSKeyException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *InvalidKMSKeyException) OrigErr() error
func (s *InvalidKMSKeyException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *InvalidKMSKeyException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *InvalidKMSKeyException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidParameterException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidParameterException) GoString() string
  // Code returns the exception type name.
func (s *InvalidParameterException) Code() string
  // Message returns the exception's message.
func (s *InvalidParameterException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *InvalidParameterException) OrigErr() error
func (s *InvalidParameterException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *InvalidParameterException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *InvalidParameterException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidS3ObjectException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s InvalidS3ObjectException) GoString() string
  // Code returns the exception type name.
func (s *InvalidS3ObjectException) Code() string
  // Message returns the exception's message.
func (s *InvalidS3ObjectException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *InvalidS3ObjectException) OrigErr() error
func (s *InvalidS3ObjectException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *InvalidS3ObjectException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *InvalidS3ObjectException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingDetection) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingDetection) GoString() string
  // SetConfidence sets the Confidence field's value.
func (s *LendingDetection) SetConfidence(v float64) *LendingDetection
  // SetGeometry sets the Geometry field's value.
func (s *LendingDetection) SetGeometry(v *Geometry) *LendingDetection
  // SetSelectionStatus sets the SelectionStatus field's value.
func (s *LendingDetection) SetSelectionStatus(v string) *LendingDetection
  // SetText sets the Text field's value.
func (s *LendingDetection) SetText(v string) *LendingDetection
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingDocument) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingDocument) GoString() string
  // SetLendingFields sets the LendingFields field's value.
func (s *LendingDocument) SetLendingFields(v []*LendingField) *LendingDocument
  // SetSignatureDetections sets the SignatureDetections field's value.
func (s *LendingDocument) SetSignatureDetections(v []*SignatureDetection) *LendingDocument
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingField) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingField) GoString() string
  // SetKeyDetection sets the KeyDetection field's value.
func (s *LendingField) SetKeyDetection(v *LendingDetection) *LendingField
  // SetType sets the Type field's value.
func (s *LendingField) SetType(v string) *LendingField
  // SetValueDetections sets the ValueDetections field's value.
func (s *LendingField) SetValueDetections(v []*LendingDetection) *LendingField
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingResult) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingResult) GoString() string
  // SetExtractions sets the Extractions field's value.
func (s *LendingResult) SetExtractions(v []*Extraction) *LendingResult
  // SetPage sets the Page field's value.
func (s *LendingResult) SetPage(v int64) *LendingResult
  // SetPageClassification sets the PageClassification field's value.
func (s *LendingResult) SetPageClassification(v *PageClassification) *LendingResult
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingSummary) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LendingSummary) GoString() string
  // SetDocumentGroups sets the DocumentGroups field's value.
func (s *LendingSummary) SetDocumentGroups(v []*DocumentGroup) *LendingSummary
  // SetUndetectedDocumentTypes sets the UndetectedDocumentTypes field's value.
func (s *LendingSummary) SetUndetectedDocumentTypes(v []*string) *LendingSummary
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LimitExceededException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LimitExceededException) GoString() string
  // Code returns the exception type name.
func (s *LimitExceededException) Code() string
  // Message returns the exception's message.
func (s *LimitExceededException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *LimitExceededException) OrigErr() error
func (s *LimitExceededException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *LimitExceededException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *LimitExceededException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LineItemFields) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LineItemFields) GoString() string
  // SetLineItemExpenseFields sets the LineItemExpenseFields field's value.
func (s *LineItemFields) SetLineItemExpenseFields(v []*ExpenseField) *LineItemFields
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LineItemGroup) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s LineItemGroup) GoString() string
  // SetLineItemGroupIndex sets the LineItemGroupIndex field's value.
func (s *LineItemGroup) SetLineItemGroupIndex(v int64) *LineItemGroup
  // SetLineItems sets the LineItems field's value.
func (s *LineItemGroup) SetLineItems(v []*LineItemFields) *LineItemGroup
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdapterVersionsInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdapterVersionsInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *ListAdapterVersionsInput) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *ListAdapterVersionsInput) SetAdapterId(v string) *ListAdapterVersionsInput
  // SetAfterCreationTime sets the AfterCreationTime field's value.
func (s *ListAdapterVersionsInput) SetAfterCreationTime(v time.Time) *ListAdapterVersionsInput
  // SetBeforeCreationTime sets the BeforeCreationTime field's value.
func (s *ListAdapterVersionsInput) SetBeforeCreationTime(v time.Time) *ListAdapterVersionsInput
  // SetMaxResults sets the MaxResults field's value.
func (s *ListAdapterVersionsInput) SetMaxResults(v int64) *ListAdapterVersionsInput
  // SetNextToken sets the NextToken field's value.
func (s *ListAdapterVersionsInput) SetNextToken(v string) *ListAdapterVersionsInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdapterVersionsOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdapterVersionsOutput) GoString() string
  // SetAdapterVersions sets the AdapterVersions field's value.
func (s *ListAdapterVersionsOutput) SetAdapterVersions(v []*AdapterVersionOverview) *ListAdapterVersionsOutput
  // SetNextToken sets the NextToken field's value.
func (s *ListAdapterVersionsOutput) SetNextToken(v string) *ListAdapterVersionsOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdaptersInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdaptersInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *ListAdaptersInput) Validate() error
  // SetAfterCreationTime sets the AfterCreationTime field's value.
func (s *ListAdaptersInput) SetAfterCreationTime(v time.Time) *ListAdaptersInput
  // SetBeforeCreationTime sets the BeforeCreationTime field's value.
func (s *ListAdaptersInput) SetBeforeCreationTime(v time.Time) *ListAdaptersInput
  // SetMaxResults sets the MaxResults field's value.
func (s *ListAdaptersInput) SetMaxResults(v int64) *ListAdaptersInput
  // SetNextToken sets the NextToken field's value.
func (s *ListAdaptersInput) SetNextToken(v string) *ListAdaptersInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdaptersOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListAdaptersOutput) GoString() string
  // SetAdapters sets the Adapters field's value.
func (s *ListAdaptersOutput) SetAdapters(v []*AdapterOverview) *ListAdaptersOutput
  // SetNextToken sets the NextToken field's value.
func (s *ListAdaptersOutput) SetNextToken(v string) *ListAdaptersOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListTagsForResourceInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListTagsForResourceInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *ListTagsForResourceInput) Validate() error
  // SetResourceARN sets the ResourceARN field's value.
func (s *ListTagsForResourceInput) SetResourceARN(v string) *ListTagsForResourceInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListTagsForResourceOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ListTagsForResourceOutput) GoString() string
  // SetTags sets the Tags field's value.
func (s *ListTagsForResourceOutput) SetTags(v map[string]*string) *ListTagsForResourceOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s NormalizedValue) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s NormalizedValue) GoString() string
  // SetValue sets the Value field's value.
func (s *NormalizedValue) SetValue(v string) *NormalizedValue
  // SetValueType sets the ValueType field's value.
func (s *NormalizedValue) SetValueType(v string) *NormalizedValue
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s NotificationChannel) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s NotificationChannel) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *NotificationChannel) Validate() error
  // SetRoleArn sets the RoleArn field's value.
func (s *NotificationChannel) SetRoleArn(v string) *NotificationChannel
  // SetSNSTopicArn sets the SNSTopicArn field's value.
func (s *NotificationChannel) SetSNSTopicArn(v string) *NotificationChannel
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s OutputConfig) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s OutputConfig) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *OutputConfig) Validate() error
  // SetS3Bucket sets the S3Bucket field's value.
func (s *OutputConfig) SetS3Bucket(v string) *OutputConfig
  // SetS3Prefix sets the S3Prefix field's value.
func (s *OutputConfig) SetS3Prefix(v string) *OutputConfig
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s PageClassification) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s PageClassification) GoString() string
  // SetPageNumber sets the PageNumber field's value.
func (s *PageClassification) SetPageNumber(v []*Prediction) *PageClassification
  // SetPageType sets the PageType field's value.
func (s *PageClassification) SetPageType(v []*Prediction) *PageClassification
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Point) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Point) GoString() string
  // SetX sets the X field's value.
func (s *Point) SetX(v float64) *Point
  // SetY sets the Y field's value.
func (s *Point) SetY(v float64) *Point
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Prediction) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Prediction) GoString() string
  // SetConfidence sets the Confidence field's value.
func (s *Prediction) SetConfidence(v float64) *Prediction
  // SetValue sets the Value field's value.
func (s *Prediction) SetValue(v string) *Prediction
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ProvisionedThroughputExceededException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ProvisionedThroughputExceededException) GoString() string
  // Code returns the exception type name.
func (s *ProvisionedThroughputExceededException) Code() string
  // Message returns the exception's message.
func (s *ProvisionedThroughputExceededException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ProvisionedThroughputExceededException) OrigErr() error
func (s *ProvisionedThroughputExceededException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *ProvisionedThroughputExceededException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *ProvisionedThroughputExceededException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s QueriesConfig) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s QueriesConfig) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *QueriesConfig) Validate() error
  // SetQueries sets the Queries field's value.
func (s *QueriesConfig) SetQueries(v []*Query) *QueriesConfig
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Query) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Query) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *Query) Validate() error
  // SetAlias sets the Alias field's value.
func (s *Query) SetAlias(v string) *Query
  // SetPages sets the Pages field's value.
func (s *Query) SetPages(v []*string) *Query
  // SetText sets the Text field's value.
func (s *Query) SetText(v string) *Query
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Relationship) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Relationship) GoString() string
  // SetIds sets the Ids field's value.
func (s *Relationship) SetIds(v []*string) *Relationship
  // SetType sets the Type field's value.
func (s *Relationship) SetType(v string) *Relationship
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ResourceNotFoundException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ResourceNotFoundException) GoString() string
  // Code returns the exception type name.
func (s *ResourceNotFoundException) Code() string
  // Message returns the exception's message.
func (s *ResourceNotFoundException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ResourceNotFoundException) OrigErr() error
func (s *ResourceNotFoundException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *ResourceNotFoundException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *ResourceNotFoundException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s S3Object) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s S3Object) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *S3Object) Validate() error
  // SetBucket sets the Bucket field's value.
func (s *S3Object) SetBucket(v string) *S3Object
  // SetName sets the Name field's value.
func (s *S3Object) SetName(v string) *S3Object
  // SetVersion sets the Version field's value.
func (s *S3Object) SetVersion(v string) *S3Object
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ServiceQuotaExceededException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ServiceQuotaExceededException) GoString() string
  // Code returns the exception type name.
func (s *ServiceQuotaExceededException) Code() string
  // Message returns the exception's message.
func (s *ServiceQuotaExceededException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ServiceQuotaExceededException) OrigErr() error
func (s *ServiceQuotaExceededException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *ServiceQuotaExceededException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *ServiceQuotaExceededException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s SignatureDetection) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s SignatureDetection) GoString() string
  // SetConfidence sets the Confidence field's value.
func (s *SignatureDetection) SetConfidence(v float64) *SignatureDetection
  // SetGeometry sets the Geometry field's value.
func (s *SignatureDetection) SetGeometry(v *Geometry) *SignatureDetection
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s SplitDocument) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s SplitDocument) GoString() string
  // SetIndex sets the Index field's value.
func (s *SplitDocument) SetIndex(v int64) *SplitDocument
  // SetPages sets the Pages field's value.
func (s *SplitDocument) SetPages(v []*int64) *SplitDocument
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentAnalysisInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentAnalysisInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *StartDocumentAnalysisInput) Validate() error
  // SetAdaptersConfig sets the AdaptersConfig field's value.
func (s *StartDocumentAnalysisInput) SetAdaptersConfig(v *AdaptersConfig) *StartDocumentAnalysisInput
  // SetClientRequestToken sets the ClientRequestToken field's value.
func (s *StartDocumentAnalysisInput) SetClientRequestToken(v string) *StartDocumentAnalysisInput
  // SetDocumentLocation sets the DocumentLocation field's value.
func (s *StartDocumentAnalysisInput) SetDocumentLocation(v *DocumentLocation) *StartDocumentAnalysisInput
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *StartDocumentAnalysisInput) SetFeatureTypes(v []*string) *StartDocumentAnalysisInput
  // SetJobTag sets the JobTag field's value.
func (s *StartDocumentAnalysisInput) SetJobTag(v string) *StartDocumentAnalysisInput
  // SetKMSKeyId sets the KMSKeyId field's value.
func (s *StartDocumentAnalysisInput) SetKMSKeyId(v string) *StartDocumentAnalysisInput
  // SetNotificationChannel sets the NotificationChannel field's value.
func (s *StartDocumentAnalysisInput) SetNotificationChannel(v *NotificationChannel) *StartDocumentAnalysisInput
  // SetOutputConfig sets the OutputConfig field's value.
func (s *StartDocumentAnalysisInput) SetOutputConfig(v *OutputConfig) *StartDocumentAnalysisInput
  // SetQueriesConfig sets the QueriesConfig field's value.
func (s *StartDocumentAnalysisInput) SetQueriesConfig(v *QueriesConfig) *StartDocumentAnalysisInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentAnalysisOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentAnalysisOutput) GoString() string
  // SetJobId sets the JobId field's value.
func (s *StartDocumentAnalysisOutput) SetJobId(v string) *StartDocumentAnalysisOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentTextDetectionInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentTextDetectionInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *StartDocumentTextDetectionInput) Validate() error
  // SetClientRequestToken sets the ClientRequestToken field's value.
func (s *StartDocumentTextDetectionInput) SetClientRequestToken(v string) *StartDocumentTextDetectionInput
  // SetDocumentLocation sets the DocumentLocation field's value.
func (s *StartDocumentTextDetectionInput) SetDocumentLocation(v *DocumentLocation) *StartDocumentTextDetectionInput
  // SetJobTag sets the JobTag field's value.
func (s *StartDocumentTextDetectionInput) SetJobTag(v string) *StartDocumentTextDetectionInput
  // SetKMSKeyId sets the KMSKeyId field's value.
func (s *StartDocumentTextDetectionInput) SetKMSKeyId(v string) *StartDocumentTextDetectionInput
  // SetNotificationChannel sets the NotificationChannel field's value.
func (s *StartDocumentTextDetectionInput) SetNotificationChannel(v *NotificationChannel) *StartDocumentTextDetectionInput
  // SetOutputConfig sets the OutputConfig field's value.
func (s *StartDocumentTextDetectionInput) SetOutputConfig(v *OutputConfig) *StartDocumentTextDetectionInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentTextDetectionOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartDocumentTextDetectionOutput) GoString() string
  // SetJobId sets the JobId field's value.
func (s *StartDocumentTextDetectionOutput) SetJobId(v string) *StartDocumentTextDetectionOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartExpenseAnalysisInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartExpenseAnalysisInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *StartExpenseAnalysisInput) Validate() error
  // SetClientRequestToken sets the ClientRequestToken field's value.
func (s *StartExpenseAnalysisInput) SetClientRequestToken(v string) *StartExpenseAnalysisInput
  // SetDocumentLocation sets the DocumentLocation field's value.
func (s *StartExpenseAnalysisInput) SetDocumentLocation(v *DocumentLocation) *StartExpenseAnalysisInput
  // SetJobTag sets the JobTag field's value.
func (s *StartExpenseAnalysisInput) SetJobTag(v string) *StartExpenseAnalysisInput
  // SetKMSKeyId sets the KMSKeyId field's value.
func (s *StartExpenseAnalysisInput) SetKMSKeyId(v string) *StartExpenseAnalysisInput
  // SetNotificationChannel sets the NotificationChannel field's value.
func (s *StartExpenseAnalysisInput) SetNotificationChannel(v *NotificationChannel) *StartExpenseAnalysisInput
  // SetOutputConfig sets the OutputConfig field's value.
func (s *StartExpenseAnalysisInput) SetOutputConfig(v *OutputConfig) *StartExpenseAnalysisInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartExpenseAnalysisOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartExpenseAnalysisOutput) GoString() string
  // SetJobId sets the JobId field's value.
func (s *StartExpenseAnalysisOutput) SetJobId(v string) *StartExpenseAnalysisOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartLendingAnalysisInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartLendingAnalysisInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *StartLendingAnalysisInput) Validate() error
  // SetClientRequestToken sets the ClientRequestToken field's value.
func (s *StartLendingAnalysisInput) SetClientRequestToken(v string) *StartLendingAnalysisInput
  // SetDocumentLocation sets the DocumentLocation field's value.
func (s *StartLendingAnalysisInput) SetDocumentLocation(v *DocumentLocation) *StartLendingAnalysisInput
  // SetJobTag sets the JobTag field's value.
func (s *StartLendingAnalysisInput) SetJobTag(v string) *StartLendingAnalysisInput
  // SetKMSKeyId sets the KMSKeyId field's value.
func (s *StartLendingAnalysisInput) SetKMSKeyId(v string) *StartLendingAnalysisInput
  // SetNotificationChannel sets the NotificationChannel field's value.
func (s *StartLendingAnalysisInput) SetNotificationChannel(v *NotificationChannel) *StartLendingAnalysisInput
  // SetOutputConfig sets the OutputConfig field's value.
func (s *StartLendingAnalysisInput) SetOutputConfig(v *OutputConfig) *StartLendingAnalysisInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartLendingAnalysisOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s StartLendingAnalysisOutput) GoString() string
  // SetJobId sets the JobId field's value.
func (s *StartLendingAnalysisOutput) SetJobId(v string) *StartLendingAnalysisOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s TagResourceInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s TagResourceInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *TagResourceInput) Validate() error
  // SetResourceARN sets the ResourceARN field's value.
func (s *TagResourceInput) SetResourceARN(v string) *TagResourceInput
  // SetTags sets the Tags field's value.
func (s *TagResourceInput) SetTags(v map[string]*string) *TagResourceInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s TagResourceOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s TagResourceOutput) GoString() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ThrottlingException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ThrottlingException) GoString() string
  // Code returns the exception type name.
func (s *ThrottlingException) Code() string
  // Message returns the exception's message.
func (s *ThrottlingException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ThrottlingException) OrigErr() error
func (s *ThrottlingException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *ThrottlingException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *ThrottlingException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UndetectedSignature) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UndetectedSignature) GoString() string
  // SetPage sets the Page field's value.
func (s *UndetectedSignature) SetPage(v int64) *UndetectedSignature
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UnsupportedDocumentException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UnsupportedDocumentException) GoString() string
  // Code returns the exception type name.
func (s *UnsupportedDocumentException) Code() string
  // Message returns the exception's message.
func (s *UnsupportedDocumentException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *UnsupportedDocumentException) OrigErr() error
func (s *UnsupportedDocumentException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *UnsupportedDocumentException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *UnsupportedDocumentException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UntagResourceInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UntagResourceInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *UntagResourceInput) Validate() error
  // SetResourceARN sets the ResourceARN field's value.
func (s *UntagResourceInput) SetResourceARN(v string) *UntagResourceInput
  // SetTagKeys sets the TagKeys field's value.
func (s *UntagResourceInput) SetTagKeys(v []*string) *UntagResourceInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UntagResourceOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UntagResourceOutput) GoString() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UpdateAdapterInput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UpdateAdapterInput) GoString() string
  // Validate inspects the fields of the type to determine if they are valid.
func (s *UpdateAdapterInput) Validate() error
  // SetAdapterId sets the AdapterId field's value.
func (s *UpdateAdapterInput) SetAdapterId(v string) *UpdateAdapterInput
  // SetAdapterName sets the AdapterName field's value.
func (s *UpdateAdapterInput) SetAdapterName(v string) *UpdateAdapterInput
  // SetAutoUpdate sets the AutoUpdate field's value.
func (s *UpdateAdapterInput) SetAutoUpdate(v string) *UpdateAdapterInput
  // SetDescription sets the Description field's value.
func (s *UpdateAdapterInput) SetDescription(v string) *UpdateAdapterInput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UpdateAdapterOutput) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s UpdateAdapterOutput) GoString() string
  // SetAdapterId sets the AdapterId field's value.
func (s *UpdateAdapterOutput) SetAdapterId(v string) *UpdateAdapterOutput
  // SetAdapterName sets the AdapterName field's value.
func (s *UpdateAdapterOutput) SetAdapterName(v string) *UpdateAdapterOutput
  // SetAutoUpdate sets the AutoUpdate field's value.
func (s *UpdateAdapterOutput) SetAutoUpdate(v string) *UpdateAdapterOutput
  // SetCreationTime sets the CreationTime field's value.
func (s *UpdateAdapterOutput) SetCreationTime(v time.Time) *UpdateAdapterOutput
  // SetDescription sets the Description field's value.
func (s *UpdateAdapterOutput) SetDescription(v string) *UpdateAdapterOutput
  // SetFeatureTypes sets the FeatureTypes field's value.
func (s *UpdateAdapterOutput) SetFeatureTypes(v []*string) *UpdateAdapterOutput
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ValidationException) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s ValidationException) GoString() string
  // Code returns the exception type name.
func (s *ValidationException) Code() string
  // Message returns the exception's message.
func (s *ValidationException) Message() string
  // OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ValidationException) OrigErr() error
func (s *ValidationException) Error() string
  // Status code returns the HTTP status code for the request's response error.
func (s *ValidationException) StatusCode() int
  // RequestID returns the service's response RequestID for request.
func (s *ValidationException) RequestID() string
  // String returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Warning) String() string
  // GoString returns the string representation.
  //
  // API parameter values that are decorated as "sensitive" in the API will not
  // be included in the string output. The member name will be present, but the
  // value will be replaced with "sensitive".
func (s Warning) GoString() string
  // SetErrorCode sets the ErrorCode field's value.
func (s *Warning) SetErrorCode(v string) *Warning
  // SetPages sets the Pages field's value.
func (s *Warning) SetPages(v []*int64) *Warning
