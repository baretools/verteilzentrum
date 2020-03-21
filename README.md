# Verteilzentrum

*A minimalistic mailing list*

Verteilzentrum is a minimalistic mailing list following the [KISS](https://en.wikipedia.org/wiki/KISS_principle) philosophy. 

**Features:**
- Single config file
- Multiple lists
- Blacklisting
- Whitelisting
- Configurable publishing rights

# Installation
0. Install golang (>=1.14), gcc, make and build-essential if you don't have them already
1. Clone the repository: `git clone https://git.bn4t.me/bn4t/verteilzentrum.git`
2. Checkout the latest stable tag 
3. Run `make build` to build the verteilzentrum binary
4. Run `sudo make install` to install verteilzentrum on your system. This will create the directories `/etc/verteilzentrum` (config directory) and `/var/lib/verteilzentrum` (data directory). Additionally the user `verteilzentrum` will be created.
5. If you have systemd installed you can run `sudo make install-systemd` to install the systemd service. Run `service verteilzentrum start` to start the verteilzentrum service. Verteilzentrum will automatically run as the `verteilzentrum` user.

You can make other certificates and private key files accessible to the `verteilzentrum` user with the following command:
````shell script
setfacl -m u:verteilzentrum:rx /etc/letsencrypt/ 
````

#### Increasing deliverability
To increase deliverability it is recommended to set up an [SPF](https://en.wikipedia.org/wiki/Sender_Policy_Framework) and [DMARC](https://en.wikipedia.org/wiki/DMARC) record. 


# How to use

#### Subscribing to a list
Send an email (content doesn't matter) to `subscribe+$list_name`. E.g. `subscribe+news@lists.example.com`.

You will receive a confirmation email that subscribing was successful.

#### Unsubscribing from a list
Send an email (content doesn't matter) to `unsubscribe+$list_name`. E.g. `unsubscribe+news@lists.example.com`.

You will receive a confirmation email that unsubscribing was successful.

# Config
## General options

##### hostname
The hostname of the list server. If you specify a TLS certificate it has to be valid for this hostname.

##### read_timeout
SMTP read timeout in milliseconds.

##### write_timeout
SMTP write timeout in milliseconds.

##### max_message_bytes
Maximum incoming message size in bytes.

##### tls_cert_file
Path to the TLS certificate file.

##### tls_key_file
Path to the corresponding private key to the TLS certificate. 

To disable inbound TLS just comment out both TLS settings.

#### Example
````toml
[verteilzentrum]
hostname = "lists.example.com"
read_timeout = 100000
write_timeout = 100000
max_message_bytes = 1048576 # 1024 * 1024
tls_cert_file = "/some/path/cert.pem"
tls_key_file = "/some/path/key.pem"
````

## Lists
Lists are represented toml tables in an array.

#### Table elements
##### name 
The name of the list which also serves as the list address.
##### whitelist
Array of whitelisted email addresses which are allowed to interact with the list. Supports wildcards.

If empty the whitelist is disabled.
##### blacklist
Array of blacklisted email addresses. 

Blacklisted addresses are not allowed to interact in any way with the list. Can be empty. Supports wildcards. 

Important: The Blacklist has a higher priority than the whitelist.
##### can_publish
Array of email addresses which are allowed to publish messages to the list. Supports wildcards.

#### Example
````toml
[[list]]
name = "news@lists.example.com"
whitelist = ["*"]
blacklist = ["baduser@gmail.com"]
can_publish = ["admin@example.com"]

[[list]]
name = "private-list@lists.example.com"
whitelist = ["postmaster@example.com","admin@example.com"]
blacklist = []
can_publish = ["admin@example.com", "postmaster@example.com"]
````


For further examples take a look at the [example config](configs/config.example.toml).

## Command line flags
- `-config <config file>` - The location of the config file to use. Defaults to `config.toml` in the working directory.
- `-datadir <data directory>` - The location where all persistent data is stored. Defaults to the working directory.

# Contributing
Feel free to send patches to me@bn4t.me or to open pull requests on Github. 

# License
This project is licensed under the GPL version 3 and later. See the [LICENSE](LICENSE) file. 