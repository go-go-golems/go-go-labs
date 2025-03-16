-- Search for all emails from January 2025
tell application "Mail"
    -- Create a temporary rule for the search
    set searchRule to make new rule with properties {name:"January 2025 Search", enabled:true}
    
    tell searchRule
        -- Set start date to January 1, 2025
        set startDate to date "1/1/2025 12:00:00 AM"
        -- Set end date to January 31, 2025
        set endDate to date "1/31/2025 11:59:59 PM"
        
        -- Add date received condition for start date
        make new rule condition with properties {rule type:date received, qualifier:greater than value, expression:startDate}
        
        -- Add date received condition for end date
        make new rule condition with properties {rule type:date received, qualifier:less than value, expression:endDate}
        
        -- Ensure both conditions must be met
        set all conditions must be met to true
        
        -- Find matching messages in inbox
        set matchingMessages to (perform mail action with messages messages of inbox for rule searchRule)
        
        -- Clean up by deleting the temporary rule
        -- delete searchRule
    end tell
    
    -- Return the matching messages
    return matchingMessages
end tell 