# keycloak-ssh-auth

A Go-based CLI tool for use with OpenSSH's `AuthorizedKeysCommand`, which fetches public SSH keys for users from a Keycloak attribute using a service account.

---

## What It Does

When configured in `sshd_config`, this tool allows you to authenticate users using SSH keys stored in Keycloak user attributes. It supports:

- Multi-valued and newline-separated SSH key attributes
- Secure access via OAuth2 client credentials
- System-level integration with OpenSSH
- Respecting environment `http_proxy` and system CAs

---

## Installation

### 1. Download


Download the current release from the [release page](https://github.com/siegy22/keycloak-ssh-auth/releases).


or build it yourself:

```bash
go build -o /usr/local/bin/keycloak-ssh-auth .
chmod +x /usr/local/bin/keycloak-ssh-auth
```

### 2. Configuration File

Create a config file at `/etc/keycloak-ssh-auth/config.yaml`:

```yaml
url: "https://keycloak.example.com"
realm: "myrealm"
client_id: "ssh-auth"
client_secret: "super-secret"
attribute: "sshPublicKey"
debug: false
```

- `url`: Base URL of your Keycloak instance
- `realm`: Name of your Keycloak realm
- `client_id`: Client ID for your service account
- `client_secret`: Client secret for your service account
- `attribute`: Name of the user attribute storing public keys
- `ignore_disabled`: Don't return any SSH keys for disabled users (default: true)
- `debug`: Enable debug logging to stderr (optional)

Ensure the file is readable by the user defined in `AuthorizedKeysCommandUser` (usually `nobody`).

---

## Keycloak Setup

### 1. Create a Service Account Client

In the Keycloak admin UI:

- Go to **Clients** ? **Create client**
  - **Client ID**: `ssh-auth` (must match the config file)
  - Click **Next**
  - **Client authentication**: On
  - **Standard flow**: Uncheck
  - **Service account roles**: Check
  - Click **Save**

### 2. Assign Permissions

- Go to the `ssh-auth` client ? **Service Account Roles**
  - Assign **realm roles**:
    - `view-users`

### 3. Create the SSH Key Attribute

In **Realm Settings** ? **User Profile** tab:

- Click **Add attribute**
  - **Name**: `sshPublicKey` (must match the config file)
  - Check **Multivalued** if needed (can be enabled later)
  - Save

### 4. Store SSH Keys on User Accounts

In **Users**:

- Click on any User
- Scroll down to: `sshPublicKey`
- Enter your public key(s)

Example:
```text
ssh-ed25519 AAAA... user@host
ssh-rsa AAAA... user@host
```

---

## SSHD Integration

### SELinux Users

If SELinux is enabled (common on RHEL, Fedora, and CentOS systems), the SSH daemon may block the `AuthorizedKeysCommand` from running as expected. You must enable the SELinux boolean that permits NSS lookups from non-privileged users:

```bash
sudo setsebool -P nis_enabled on
```

This allows the `nobody` user to perform the necessary name service lookups (e.g., DNS, LDAP, HTTP) required to query Keycloak.


Edit `/etc/ssh/sshd_config`:

```conf
AuthorizedKeysCommand /usr/local/bin/keycloak-ssh-auth
AuthorizedKeysCommandUser nobody
```

> **Note**: `AuthorizedKeysCommand` can't take arguments, so all config must come from the YAML file.

Restart SSH:

```bash
sudo systemctl restart sshd
```

---

## Verifying

Manually test the command:

```bash
sudo -u nobody /usr/local/bin/keycloak-ssh-auth yourusername
```

It should print one or more public keys.

To debug:

```bash
sudo /usr/sbin/sshd -d -p 2222
ssh -p 2222 yourusername@localhost
```

---

## Security Notes

- Your config file contains secrets secure it:
  ```bash
  chmod 640 /etc/keycloak-ssh-auth/config.yaml
  chown root:nobody /etc/keycloak-ssh-auth/config.yaml
  ```
- Ensure the binary is not writable by unauthorized users.
- Optionally set `http_proxy` in the environment if you're behind a proxy.
