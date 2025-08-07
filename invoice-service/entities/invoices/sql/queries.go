package sql

import _ "embed"

// Invoice SQL queries
//
//go:embed scripts/create_invoice.sql
var CreateInvoiceQuery string

//go:embed scripts/get_invoice_by_id.sql
var GetInvoiceByIDQuery string

//go:embed scripts/get_invoice_by_number.sql
var GetInvoiceByNumberQuery string

//go:embed scripts/list_invoices.sql
var ListInvoicesQuery string

//go:embed scripts/update_invoice.sql
var UpdateInvoiceQuery string

//go:embed scripts/delete_invoice.sql
var DeleteInvoiceQuery string

//go:embed scripts/count_invoices.sql
var CountInvoicesQuery string

// Invoice Details SQL queries
//
//go:embed scripts/create_invoice_detail.sql
var CreateInvoiceDetailQuery string

//go:embed scripts/get_invoice_detail_by_id.sql
var GetInvoiceDetailByIDQuery string

//go:embed scripts/get_invoice_details_by_invoice_id.sql
var GetInvoiceDetailsByInvoiceIDQuery string

//go:embed scripts/list_invoice_details.sql
var ListInvoiceDetailsQuery string

//go:embed scripts/update_invoice_detail.sql
var UpdateInvoiceDetailQuery string

//go:embed scripts/delete_invoice_detail.sql
var DeleteInvoiceDetailQuery string

//go:embed scripts/get_invoice_total_from_details.sql
var GetInvoiceTotalFromDetailsQuery string

//go:embed scripts/update_invoice_total.sql
var UpdateInvoiceTotalQuery string 