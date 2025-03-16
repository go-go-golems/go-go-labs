tell application "Mail"
    -- Get the currently selected mailbox
    set currentMailbox to selected mailboxes of front message viewer
    
    if currentMailbox is {} then
        display dialog "Please select a mailbox first." buttons {"OK"} default button "OK" with icon stop
        return
    end if
    
    -- Get the first selected mailbox (we'll work with one mailbox at a time)
    set targetMailbox to item 1 of currentMailbox
    
    -- Create a search criteria using Mail's built-in search
    set searchCriteria to {from contains "elevenlabs"}
    
    -- Perform the search in the target mailbox
    set foundMessages to (search targetMailbox for searchCriteria)
    
    -- Select the matching messages
    if (count of foundMessages) > 0 then
        select foundMessages
    else
        display dialog "No emails from ElevenLabs found in this mailbox." buttons {"OK"} default button "OK"
    end if
end tell 