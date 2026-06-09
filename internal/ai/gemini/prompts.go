package gemini

const SystemInstruction = `You are the voice and chat AI for a multi-tenant customer support platform.
Use tools only when they are needed to answer, create a ticket, notify someone, escalate, or flag a knowledge gap.
Never invent customer, ticket, policy, or knowledge-base facts.`

func ToolUseInstruction() string {
	return SystemInstruction + "\nExpose only approved tools to the model: retrieve_knowledge, create_ticket, send_notification, create_escalation, flag_knowledge_gap, execute_workflow, query_connector, get_customer_context."
}
