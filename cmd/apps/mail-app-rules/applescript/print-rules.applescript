tell application "Mail"
	-- Get all rules
	set allRules to every rule
	
	set ruleCount to count of allRules
	log "Found " & ruleCount & " rules"
	
	repeat with i from 1 to ruleCount
		set currentRule to item i of allRules
		
		-- Log basic rule information
		log "---------------------------------------"
		log "Rule " & i & ": " & name of currentRule
		log "Enabled: " & enabled of currentRule
		
		-- Log whether all conditions must be met
		if all conditions must be met of currentRule then
			log "Logic: All conditions must be met (AND)"
		else
			log "Logic: Any condition can be met (OR)"
		end if
		
		-- Get rule conditions
		set ruleConditions to rule conditions of currentRule
		set conditionCount to count of ruleConditions
		log "Number of conditions: " & conditionCount
		
		-- Log each condition
		repeat with j from 1 to conditionCount
			set currentCondition to item j of ruleConditions
			log "  Condition " & j & ":"
			log "    Header: " & header of currentCondition
			log "    Type: " & rule type of currentCondition
			log "    Qualifier: " & qualifier of currentCondition
			log "    Expression: " & expression of currentCondition
		end repeat
		
		-- Log actions
		log "Actions:"
		
		-- Color action
		if color message of currentRule is not "none" then
			log "  Color messages: " & color message of currentRule
		end if
		
		-- Delete action
		if delete message of currentRule then
			log "  Delete messages"
		end if
		
		
		-- Copy action
		try
			set copyMailbox to copy message of currentRule
			log "  Copy messages to: " & name of copyMailbox
		on error
			-- No copy action
		end try
		
		-- Mark read action
		if mark read of currentRule then
			log "  Mark as read"
		end if
		
		-- Mark flagged action
		if mark flagged of currentRule then
			log "  Mark as flagged"
		end if
		
		-- Flag color action
		set flagIndex to mark flag index of currentRule
		if flagIndex is not -1 then
			log "  Set flag index to: " & flagIndex
		end if
		
		-- Forward action
		if forward message of currentRule is not "" then
			log "  Forward to: " & forward message of currentRule
			log "  Forward text: " & forward text of currentRule
		end if
		
		-- Redirect action
		if redirect message of currentRule is not "" then
			log "  Redirect to: " & redirect message of currentRule
		end if
		
		-- Reply action
		if reply text of currentRule is not "" then
			log "  Auto-reply with text: " & reply text of currentRule
		end if
		
		-- Run script action
		try
			set scriptFile to run script of currentRule
			log "  Run script: " & scriptFile
		on error
			-- No script action
		end try
		
		-- Stop evaluating rules
		if stop evaluating rules of currentRule then
			log "  Stop evaluating rules after this rule"
		end if
	end repeat
	
	log "---------------------------------------"
	log "End of rule analysis"
end tell