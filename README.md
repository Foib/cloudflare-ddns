# cloudflare-ddns

DDNS Client for Cloudflare

## Usage
    cloudflare-ddns -config path-to-config

## Flags
    -h
        Show help
    -comment string
        Comment about the DNS Record
    -config string
        Path to the config file
    -email string
        Cloudflare API Email
    -force
        Force the first update to Cloudflare
    -generate
        Generate a config file with the provided parameters
    -interval int
        Interval in seconds to check for IP changes (default 600)
    -key string
        Cloudflare API Key
    -name string
        Cloudflare DNS Record Name
    -proxied
        Whether the DNS Record is proxied by Cloudflare or not
    -record string
        Cloudflare DNS Record ID
    -ttl int
        TTL of the DNS Record (default 3600)
    -zone string
        Cloudflare Zone ID

## Build

    go build