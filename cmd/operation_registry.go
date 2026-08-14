package cmd

func operationBindings() map[string]operationBinding {
	definitions := []operationBinding{
		op("account create-sub", "create-sub-account", "connect", "account-center", "POST", "/v1/accounts/create_accounts"),
		op("account additional-documents", "get-additional-documents", "connect", "account-center", "GET", "/v1/accounts/get_additional"),
		op("account create", "create-account", "connect", "account-center", "POST", "/v1/accounts"),
		op("account list", "list-connected-accounts", "connect", "account-center", "GET", "/v1/accounts"),
		op("account get", "retrieve-account", "connect", "account-center", "GET", "/v1/accounts/{id}"),
		op("rfi list", "list-rfis", "connect", "account-center", "GET", "/v1/rfis"),
		op("rfi get", "retrieve-rfi", "connect", "account-center", "GET", "/v1/rfis/{id}"),
		opWithFields("rfi answer", "answer-rfi", "connect", "account-center", "POST", "/v1/rfis/answer",
			requiredField("rfi_id", "string"), requiredField("answer", "array")),

		opWithFields("file upload", "upload-file", "supporting", "account-center", "POST", "/v1/files/upload",
			requiredField("file", "string")),
		opWithFields("file download-links", "download-links", "supporting", "account-center", "POST", "/v1/files/download_links",
			requiredField("file_ids", "array")),

		op("issuing card create", "create-card", "issuing", "card-issuance", "POST", "/v1/issuing/cards"),
		op("issuing card list", "list-cards", "issuing", "card-issuance", "GET", "/v1/issuing/cards"),
		op("issuing card update", "update-card", "issuing", "card-issuance", "POST", "/v1/issuing/cards/{id}"),
		op("issuing card get", "retrieve-card", "issuing", "card-issuance", "GET", "/v1/issuing/cards/{id}"),
		op("issuing card update-status", "update-card-status", "issuing", "card-issuance", "POST", "/v1/issuing/cards/{id}/status"),
		op("issuing card get-secure", "retrieve-card-secure", "issuing", "card-issuance", "GET", "/v1/issuing/cards/{id}/secure"),
		op("issuing card iframe-url", "create-pan-token", "issuing", "card-issuance", "POST", "/v1/issuing/cards/{id}/token"),
		op("issuing card recharge", "card-recharge", "issuing", "card-issuance", "POST", "/v1/issuing/cards/{id}/recharge"),
		op("issuing card withdraw", "card-withdraw", "issuing", "card-issuance", "POST", "/v1/issuing/cards/{id}/withdraw"),
		opWithFields("issuing card elevate-limit", "elevate-card-limit", "issuing", "card-issuance", "POST", "/v1/issuing/cards/{id}/elevate_limit",
			requiredField("limit_amount", "number"), optionalField("duration_in_days", "integer")),
		op("issuing card get-order", "retrieve-card-order", "issuing", "card-issuance", "GET", "/v1/issuing/cards/{id}/order"),
		opWithFields("issuing card enroll-network-protection", "enroll-card-network-protection", "issuing", "card-issuance", "POST", "/v1/issuing/cards/{id}/risk",
			requiredField("risk_control", "string"), requiredField("action_code", "string")),
		destructiveOpWithFields("issuing card remove-network-protection", "remove-card-network-protection", "issuing", "card-issuance", "DELETE", "/v1/issuing/cards/{id}/risk",
			requiredField("risk_control", "string")),
		op("issuing card activate", "activate-card", "issuing", "card-issuance", "POST", "/v1/issuing/cards/activate"),
		op("issuing card set-pin", "reset-pin", "issuing", "card-issuance", "POST", "/v1/issuing/cards/pin"),
		opWithFields("issuing card manage-pin", "manage-card-pin", "issuing", "card-issuance", "POST", "/v1/issuing/cards/manage/pin",
			requiredField("card_id", "string"), requiredField("type", "string"), requiredField("pin", "string"), optionalField("old_pin", "string")),
		op("issuing card assign", "assign-card", "issuing", "card-issuance", "POST", "/v1/issuing/cards/assign"),
		op("issuing card list-arts", "list-card-arts", "issuing", "card-issuance", "GET", "/v1/issuing/cards/arts"),
		opWithFields("issuing card set-default-art", "set-default-card-art", "issuing", "card-issuance", "POST", "/v1/issuing/cards/arts/default",
			requiredField("card_art_id", "string")),
		op("issuing merchant-brand", "list-merchant-brands", "issuing", "card-issuance", "GET", "/v1/issuing/merchant_brands"),
		op("issuing cardholder create", "create-cardholder", "issuing", "card-issuance", "POST", "/v1/issuing/cardholders"),
		op("issuing cardholder list", "list-cardholders", "issuing", "card-issuance", "GET", "/v1/issuing/cardholders"),
		op("issuing cardholder update", "update-cardholder", "issuing", "card-issuance", "POST", "/v1/issuing/cardholders/{id}"),
		op("issuing cardholder get", "retrieve-cardholder", "issuing", "card-issuance", "GET", "/v1/issuing/cardholders/{id}"),
		op("issuing transaction list", "list-cards-transactions", "issuing", "card-issuance", "GET", "/v1/issuing/transactions"),
		op("issuing transaction get", "retrieve-cards-transaction", "issuing", "card-issuance", "GET", "/v1/issuing/transactions/{id}"),
		opWithFields("issuing transaction claim-unsolicited-refund", "claim-unsolicited-refund", "issuing", "card-issuance", "POST", "/v1/issuing/transactions/unsolicited_refund/release",
			requiredField("related_transaction_id", "string"), optionalField("remark", "string")),
		op("issuing product list", "list-card-products", "issuing", "card-issuance", "GET", "/v1/issuing/products"),
		op("issuing transfer create", "create-issuing-transfer", "issuing", "card-issuance", "POST", "/v1/issuing/transfers"),
		op("issuing transfer get", "retrieve-issuing-transfer", "issuing", "card-issuance", "GET", "/v1/issuing/transfers/{id}"),
		opWithFields("issuing balance get", "retrieve-issuing-balance", "issuing", "card-issuance", "POST", "/v1/issuing/balances",
			requiredField("currency", "string")),
		op("issuing balance list", "list-issuing-balances", "issuing", "card-issuance", "GET", "/v1/issuing/balances"),
		op("issuing balance transactions", "list-issuing-balances-transactions", "issuing", "card-issuance", "GET", "/v1/issuing/balances/transactions"),
		op("issuing report download", "download-report-file", "issuing", "card-issuance", "GET", "/v1/issuing/reports/{id}"),
		op("issuing report create", "create-report", "issuing", "card-issuance", "POST", "/v1/issuing/reports"),
		op("simulate authorization", "simulate-authorization", "issuing", "card-issuance", "POST", "/v1/simulation/issuing/authorization"),
		opWithFields("simulate reversal", "simulate-reversal", "issuing", "card-issuance", "POST", "/v1/simulation/issuing/reversal",
			requiredField("transaction_id", "string")),

		op("banking transfer list", "list-transfers", "banking", "global-account", "GET", "/v1/transfer"),
		op("banking transfer create", "create-transfer", "banking", "global-account", "POST", "/v1/transfer"),
		op("banking transfer get", "retrieve-transfer", "banking", "global-account", "GET", "/v1/transfer/{id}"),
		op("banking payout create", "create-payout", "banking", "global-account", "POST", "/v1/payouts"),
		op("banking payout list", "list-payouts", "banking", "global-account", "GET", "/v1/payouts"),
		op("banking payout get", "retrieve-payout", "banking", "global-account", "GET", "/v1/payouts/{id}"),
		op("banking beneficiary create", "create-beneficiary", "banking", "global-account", "POST", "/v1/beneficiaries"),
		op("banking beneficiary list", "list-beneficiaries", "banking", "global-account", "GET", "/v1/beneficiaries"),
		op("banking beneficiary get", "retrieve-beneficiary", "banking", "global-account", "GET", "/v1/beneficiaries/{id}"),
		op("banking beneficiary update", "update-beneficiary", "banking", "global-account", "POST", "/v1/beneficiaries/{id}"),
		destructiveOp("banking beneficiary delete", "delete-beneficiary", "banking", "global-account", "POST", "/v1/beneficiaries/{id}/delete"),
		op("banking beneficiary payment-methods", "list-payment-methods", "banking", "global-account", "GET", "/v1/beneficiaries/paymentmethods"),
		op("banking beneficiary check", "check-beneficiary", "banking", "global-account", "POST", "/v1/beneficiaries/check"),
		op("banking balance get", "retrieve-balance", "banking", "global-account", "GET", "/v1/balances/{currency}"),
		op("banking balance list", "list-balances", "banking", "global-account", "GET", "/v1/balances"),
		op("banking balance transactions", "list-balances-transactions", "banking", "global-account", "GET", "/v1/balances/transactions"),
		op("banking deposit list", "list-deposits", "banking", "global-account", "GET", "/v1/deposit"),
		op("banking deposit get", "retrieve-deposit", "banking", "global-account", "GET", "/v1/deposit/{id}"),
		op("banking virtual-account list", "list-virtual-accounts", "banking", "global-account", "GET", "/v1/virtual/accounts"),
		op("banking virtual-account create", "create-virtual-account", "banking", "global-account", "POST", "/v1/virtual/accounts"),
		op("banking virtual-account application list", "list-virtual-account-applications", "banking", "global-account", "GET", "/v1/virtual/applications"),
		op("banking virtual-account application retrieve", "retrieve-virtual-account-application", "banking", "global-account", "GET", "/v1/virtual/applications/{application_id}"),
		op("banking conversion list", "list-conversion", "banking", "global-account", "GET", "/v1/conversion"),
		op("banking conversion create", "create-conversion", "banking", "global-account", "POST", "/v1/conversion"),
		op("banking conversion get", "retrieve-conversion", "banking", "global-account", "GET", "/v1/conversion/{id}"),
		op("banking conversion dates", "list-conversion-dates", "banking", "global-account", "GET", "/v1/conversion/conversion_dates"),
		op("banking conversion quote", "create-quote", "banking", "global-account", "POST", "/v1/conversion/quote"),
		op("banking exchange-rate list", "list-current-rates", "banking", "global-account", "GET", "/v1/exchange/rates"),
		op("simulate deposit", "simulate-deposit-creation", "banking", "global-account", "POST", "/v1/simulation/deposit"),

		op("payment intent create", "create-payment-intent", "payment", "global-acquiring", "POST", "/v2/payment_intents/create"),
		op("payment intent get", "retrieve-payment-intent", "payment", "global-acquiring", "GET", "/v2/payment_intents/{id}"),
		op("payment intent update", "update-payment-intent", "payment", "global-acquiring", "POST", "/v2/payment_intents/{id}"),
		op("payment intent confirm", "confirm-payment-intent", "payment", "global-acquiring", "POST", "/v2/payment_intents/{id}/confirm"),
		op("payment intent capture", "capture-payment-intent", "payment", "global-acquiring", "POST", "/v2/payment_intents/{id}/capture"),
		destructiveOp("payment intent cancel", "cancel-payment-intent", "payment", "global-acquiring", "POST", "/v2/payment_intents/{id}/cancel"),
		op("payment intent list", "list-payment-intents", "payment", "global-acquiring", "GET", "/v2/payment_intents"),
		op("payment attempt get", "retrieve-payment-attempt", "payment", "global-acquiring", "GET", "/v2/payment/payment_attempts/{id}"),
		op("payment attempt list", "list-payment-attempts", "payment", "global-acquiring", "GET", "/v2/payment/payment_attempts"),
		op("payment refund create", "create-refund", "payment", "global-acquiring", "POST", "/v2/payment/refunds"),
		op("payment refund list", "list-refunds", "payment", "global-acquiring", "GET", "/v2/payment/refunds"),
		op("payment refund get", "retrieve-refund", "payment", "global-acquiring", "GET", "/v2/payment/refunds/{id}"),
		op("payment settlement list", "get-list-of-settlements", "payment", "global-acquiring", "GET", "/v2/payment/settlements"),
		op("payment balance get", "retrieve-balance", "payment", "global-acquiring", "GET", "/v2/payment/balances/{currency}"),
		op("payment balance list", "list-balances", "payment", "global-acquiring", "GET", "/v2/payment/balances"),
		op("payment payout create", "create-payout", "payment", "global-acquiring", "POST", "/v2/payment/payout/create"),
		op("payment payout get", "retrieve-payout", "payment", "global-acquiring", "GET", "/v2/payment/payout/{payout_id}"),
		op("payment payout list", "list-payouts", "payment", "global-acquiring", "GET", "/v2/payment/payout"),
		op("payment bank-account create", "create-bank-account", "payment", "global-acquiring", "POST", "/v2/payment/bankaccount/create"),
		op("payment bank-account get", "retrieve-bank-account", "payment", "global-acquiring", "GET", "/v2/payment/bankaccount/{id}"),
		op("payment bank-account update", "update-bank-account", "payment", "global-acquiring", "POST", "/v2/payment/bankaccount/{id}"),
		op("payment bank-account list", "list-bank-accounts", "payment", "global-acquiring", "GET", "/v2/payment/bankaccount"),
		op("payment terminal register", "register-terminal", "payment", "global-acquiring", "POST", "/v2/terminal/register"),
		op("payment terminal get-pin-key", "get-pin-key", "payment", "global-acquiring", "POST", "/v2/terminal/getPinKey"),
	}

	bindings := make(map[string]operationBinding, len(definitions))
	for _, definition := range definitions {
		if _, exists := bindings[definition.Command]; exists {
			panic("duplicate operation registry command: " + definition.Command)
		}
		bindings[definition.Command] = definition
	}
	return bindings
}

func op(command, operationID, domain, docsScope, method, path string) operationBinding {
	risk := riskWrite
	if method == "GET" {
		risk = riskRead
	}
	return operationBinding{
		Command: command, OperationID: operationID, Domain: domain, DocsScope: docsScope,
		Method: method, Path: path, Risk: risk,
	}
}

func opWithFields(command, operationID, domain, docsScope, method, path string, fields ...manifestBodyField) operationBinding {
	binding := op(command, operationID, domain, docsScope, method, path)
	binding.FixedBodyFields = fields
	return binding
}

func destructiveOp(command, operationID, domain, docsScope, method, path string) operationBinding {
	binding := op(command, operationID, domain, docsScope, method, path)
	binding.Risk = riskDestructive
	return binding
}

func destructiveOpWithFields(command, operationID, domain, docsScope, method, path string, fields ...manifestBodyField) operationBinding {
	binding := destructiveOp(command, operationID, domain, docsScope, method, path)
	binding.FixedBodyFields = fields
	return binding
}

func requiredField(name, fieldType string) manifestBodyField {
	return manifestBodyField{Name: name, Type: fieldType, Required: true}
}

func optionalField(name, fieldType string) manifestBodyField {
	return manifestBodyField{Name: name, Type: fieldType, Required: false}
}
