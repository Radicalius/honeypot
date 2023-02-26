module ftp_honeypot

go 1.15

require (
    logging v1.0.0
    reporting v1.0.0
)

replace logging => ../../packages/logging
replace reporting => ../../packages/reporting