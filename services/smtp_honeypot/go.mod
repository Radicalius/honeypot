module smtp_honeypot

go 1.15

require (
    logging v0.0.0-00010101000000-000000000000
	reporting v1.0.0
)

replace logging => ../../packages/logging
replace reporting => ../../packages/reporting