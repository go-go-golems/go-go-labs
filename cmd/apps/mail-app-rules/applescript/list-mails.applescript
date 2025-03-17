tell application "Mail"
	-- Get first 10 messages from inbox
	set inboxMessages to (get first 10 messages of inbox)
	-- Determine how many messages to process (up to 10)
	set messageCount to count of inboxMessages
	if messageCount > 10 then
		set messageCount to 10
	end if
	
	log "Logging details of first " & messageCount & " messages in inbox:"
	log "---------------------------------------"
	
	-- Loop through the first 10 messages (or fewer if inbox has less)
	repeat with i from 1 to messageCount
		set currentMessage to item i of inboxMessages
		
		log "MESSAGE " & i & ":"
		
		-- Log basic message information
		log "   Subject: " & subject of currentMessage
		log "   From: " & sender of currentMessage
		
		-- Get recipients (To field)
		set recipientList to ""
		try
			set toRecipients to to recipients of currentMessage
			repeat with eachRecipient in toRecipients
				set recipientList to recipientList & address of eachRecipient & ", "
			end repeat
			if length of recipientList > 0 then
				set recipientList to rich text 1 thru -3 of recipientList
			end if
		on error
			set recipientList to "None or cannot retrieve"
		end try
		log "   To: " & recipientList
		
		-- Log date information
		log "   Date Received: " & date received of currentMessage
		log "   Date Sent: " & date sent of currentMessage
		
		-- Log message size
		log "   Size: " & message size of currentMessage & " bytes"
		
		-- Log interesting headers
		log "   HEADERS OF INTEREST:"
		
		-- Try to get the Message-ID header
		try
			set messageHeaders to headers of currentMessage
			repeat with eachHeader in messageHeaders
				set headerName to name of eachHeader
				set headerContent to content of eachHeader
				
				-- Log specific headers of interest
				if headerName is "Message-ID" or headerName is "X-Mailer" or headerName is "X-Priority" or headerName is "X-Spam-Status" or headerName is "Received-SPF" or headerName is "X-Original-To" or headerName is "DKIM-Signature" or headerName is "List-ID" or headerName is "X-Forwarded-For" then
					log "      " & headerName & ": " & headerContent
				end if
			end repeat
		on error errMsg
			log "      Error getting headers: " & errMsg
		end try
		
		-- Log attachment info if any
		try
			set msgAttachments to mail attachments of currentMessage
			if (count of msgAttachments) > 0 then
				log "   ATTACHMENTS:"
				repeat with eachAttachment in msgAttachments
					log "      " & name of eachAttachment & " (" & MIME type of eachAttachment & ", " & file size of eachAttachment & " bytes)"
				end repeat
			end if
		on error
			-- No attachments or error getting them
		end try
		
		log "---------------------------------------"
	end repeat
end tell