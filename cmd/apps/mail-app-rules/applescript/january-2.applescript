log "Starting to create mail rule..."

tell application "Mail"
    try
        -- Create a new rule with initial properties
        log "Creating new rule with properties..."
        set newRule to make new rule at end of rules with properties {name:"Test", enabled:true}
        log "Rule created with initial properties"
        
        -- Add conditions within the rule context
        tell newRule
            -- First condition: DateReceived header key
            log "Adding first date condition..."
            make new rule condition at end of rule conditions with properties {Â
                header:"DateReceived", Â
                rule type:header key, Â
                qualifier:less than value, Â
                expression:"7"}
            
            -- Second condition: From header
            log "Adding second condition..."
            make new rule condition at end of rule conditions with properties {Â
                header:"", Â
                rule type:from header, Â
                qualifier:equal to value, Â
                expression:"ljldkjslfdkjsdf"}
            
            -- Set all conditions must be met
            set all conditions must be met to true
            
            -- Set actions
            -- Copy to Drafts mailbox
            tell application "Mail"
                set draftsMailbox to mailbox "Drafts" of account 1
                tell newRule
                set copy message to draftsMailbox
                end
            end tell
            
            -- Other properties (confirming defaults match the output)
            set color message to blue
            set delete message to false
            set mark read to false
            set mark flagged to false
            set mark flag index to -1
            -- set forward message to ""
            -- set redirect message to ""
            -- set reply text to none
            set stop evaluating rules to false
            
            log "All conditions and actions set successfully"
        end tell
        
        log "Rule creation completed successfully!"
    on error errMsg
        log "Error occurred: " & errMsg
        error errMsg
    end try
    
    -- You could also have the rule move messages to a specific mailbox:
    -- Uncomment the next lines to activate this functionality
    -- Create a mailbox specifically for these messages
    -- set janMailbox to make new mailbox with properties {name:"January 2025"}
    -- set transfer message of newRule to janMailbox
end tell
