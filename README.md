# Wireguard API Documentation
# API Documentation

Documentation, you can check out the [Postman API Documentation](https://documenter.getpostman.com/view/13213405/2sAYJ4hfaL).


# Requirements for the Service

1. **Ubuntu/Debian Distributions**: The service is designed to run on Ubuntu 24-lts
2. **Go Version**: Requires Go version 1.25 or higher.
3. **Root Credentials**: The service must be started with root privileges.

---

**Test coverage: 80.0% of statements**

---

## Quick start
1. **Download the Project**: Clone or download the project to a machine running the appropriate Linux distribution.
2. **Go to working directory**
3. **Compile the binary** go build -o wireguard-rest main.go (golang should be installed before)
4. **Copy files to server where your service will work** wireguard-rest,wireguard_api.cfg, wireguard-rest.service, start.sh
5. **Go to directory there copied files and enter command**: sudo sh start.sh
6. **Check service command**: sudo systemctl status wireguard-rest.service

---

## Endpoints

### 1. Create Server Certificates and Interface

- **Method**: `POST`
- **URL**: `http://127.0.0.1:8888/interface/new`
- **Authorization**: Bearer Token

#### Request Body

```json
{
  "ifname": "test",
  "ip": "192.168.32.1/24",
  "endpoint": "192.168.10.157",
  "port": 1002
}
```

#### Description

- **ifname**: *Interface name* — a unique name for the interface used to identify it on the server.
- **ip**: *Subnet of the interface* — the subnet in `IP/subnet mask` format.
- **endpoint**: *IP address/DNS name* — reachable from the internet for client connections.
- **port**: *Unique port number* — open on the server to accept connections.

#### Example Response

```json
{
  "result": {
    "private": "+GUCy3KidtNtcSbw/ZaTQ9xBOaNjlabh2cwgswCtakA=",
    "public": "XBKO2OEl4EUEtU7Fx5yTXbvVud2pAZsHCTd49Abuq1A=",
    "endpoint": "192.168.10.157",
    "ipmask": "192.168.32.1/24",
    "config": "[Interface]\nPrivateKey = +GUCy3KidtNtcSbw/ZaTQ9xBOaNjlabh2cwgswCtakA=\nAddress = 192.168.32.1/24\nListenPort = 1002\n",
    "ifname": "test",
    "port": 1002
  }
}
```

---

### 2. Delete Server Interface and сertificates

- **Method**: `DELETE`
- **URL**: `http://127.0.0.1:8888/interface`
- **Authorization**: Bearer Token

#### Request Body

```json
{
  "ifname": "test",
  "private": "kMRpQWTwBGjaSQIMW0V4RV3QCq3r8BgDabcv1RDUXkc="
}
```

#### Description

- Shut down the interface.
- Remove all certificates linked to the interface, including client certificates and private keys.

#### Example Response

```json
{
  "result": "ok"
}
```

---

### 3. Get Deleted Server Certificates

- **Method**: `GET`
- **URL**: `http://127.0.0.1:8888/interface/archive`
- **Authorization**: Bearer Token

#### Example Response

```json
{
  "result": [
    {
      "private": "4NH9oIIWNh7fLuv4xF9PQVJG7+8BY/Ls2/ErokK+kmI=",
      "public": "iYEnQuh7gQkkEaWUqSO3JrOA42cln6kKePQrJLOG7ic=",
      "endpoint": "192.168.10.157",
      "ipmask": "192.168.32.1/24",
      "config": "[Interface]\nPrivateKey = 4NH9oIIWNh7fLuv4xF9PQVJG7+8BY/Ls2/ErokK+kmI=\nAddress = 192.168.32.1/24\nListenPort = 1002\n",
      "ifname": "test",
      "port": 1002
    }
  ]
}
```

---

### 4. Get interfaces list

- **Method**: `GET`
- **URL**: `https://127.0.0.1:8888/interface/all`
- **Authorization**: Bearer Token

#### Example Response

```json
{
  "result": [
    {
      "ifname": "test",
      "ip": "192.168.32.1/24",
      "port": 9999,
      "private": "6E2TMoWaadmPHgnxDf+PUP+liMhWgz/KrPiJRajpq0E=",
      "public": "TEnNkVFLvKzXyJV/9yaLlg1QcR+GA8+slyAC+1NvU2s=",
      "endpoint": "10.19.44.251",
      "config": ""
    }
  ]
}
```

---



### 5. Up interface

- **Method**: `POST`
- **URL**: `http://127.0.0.1:8888/interface/start`
- **Authorization**: Bearer Token

#### Request Body

```json
{
    "ifname": "test"
}
```

#### Example Response

```json
{
  "result": "ok"
}
```
---


### 6. Shutdown interface

- **Method**: `POST`
- **URL**: `http://127.0.0.1:8888/interface/stop`
- **Authorization**: Bearer Token

#### Request Body

```json
{
    "ifname": "test"
}
```

#### Example Response

```json
{
  "result": "ok"
}
```

---



### 7. Create New Client Certificate

- **Method**: `POST`
- **URL**: `http://127.0.0.1:8888/clients/new`
- **Authorization**: Bearer Token

#### Request Body

```json
{
  "ifname": "test",
  "ip": "",
  "alloweip": "",
  "dns": "8.8.8.8"
}
```

#### Example Response

```json
{
  "result": {
    "private": "aIxapYi1RC5eZVAZEGAeV5VHdWW4coWCeF0woGE58kQ=",
    "public": "ZaKCjAUIvDtYg8BmGOXLk6GPowDIAwoz0qN8eLt8/3w=",
    "config": "[Interface]\nPrivateKey = aIxapYi1RC5eZVAZEGAeV5VHdWW4coWCeF0woGE58kQ=\nAddress = 192.168.32.2/24\nDNS = 8.8.8.8\n[Peer]\nPublicKey = njscYaHsusSQS77m2oVHN/kaooAaqGOTljOcYZicu38=\nAllowedIPs = 192.168.32.0/24\nEndpoint = 192.168.10.157:1002\nPersistentKeepalive = 20\n",
    "ip": "192.168.32.2/24"
  }
}
```

---

### 8. Delete Client Certificate

- **Method**: `DELETE`
- **URL**: `http://127.0.0.1:8888/clients`
- **Authorization**: Bearer Token

#### Request Body

```json
{
  "public": "ZaKCjAUIvDtYg8BmGOXLk6GPowDIAwoz0qN8eLt8/3w="
}
```

#### Example Response

```json
{
  "result": "ok"
}
```

---

### 9. Get Deleted Client Certificates

- **Method**: `GET`
- **URL**: `http://127.0.0.1:8888/clients/archive`
- **Authorization**: Bearer Token

#### Example Response

```json
{
  "result": [
    {
      "ifname": "test",
      "private": "gNN7nqjzrhP/grp1vehgtLPuiRZaZeiAPVxOyWJJDkU=",
      "public": "MrzADHcAwti6XeM/4ZYauQCQy2Dlq5TI0J+D6PAvOS4=",
      "ip": "192.168.32.2/24",
      "config": "[Interface]\nPrivateKey = gNN7nqjzrhP/grp1vehgtLPuiRZaZeiAPVxOyWJJDkU=\nAddress = 192.168.32.2/24\nDNS = 8.8.8.8\n[Peer]\nPublicKey = rZ39vmConnxWABmYOZWV1ufOh+NBr3KgQvxUFMB7C0k=\nAllowedIPs = 192.168.32.0/24\nEndpoint = 192.168.10.157:1002\nPersistentKeepalive = 20\n"
    }
  ]
}
```

---

### 10. Get Connection Status

- **Method**: `GET`
- **URL**: `http://127.0.0.1:8888/clients/status`
- **Authorization**: Bearer Token

#### Example Response

```json
{
  "result": [
    {
      "ifname": "test",
      "status": [
        {
          "public": "ZaKCjAUIvDtYg8BmGOXLk6GPowDIAwoz0qN8eLt8/3w=",
          "handshake": "0001-01-01T00:00:00Z",
          "reciev": 0,
          "trasmit": 0,
          "allowedip": [
            {
              "ip": "192.168.32.2",
              "mask": 32
            }
          ]
        }
      ]
    }
  ]
}
```

---

### 11. Get All Client Certificates

- **Method**: `GET`
- **URL**: `http://127.0.0.1:8888/clients/getall`
- **Authorization**: Bearer Token

#### Example Response

```json
{
  "result": [
    {
      "ifname": "test",
      "private": "6Nu2payFq/fhwYolzFY1o3nXJZwq0+BkxGmoP10Uu3I=",
      "public": "VFslwVjYebt0+vsjYiLE5kNP6f6E2eJhwQSzNCLOrFs=",
      "ip": "192.168.32.3/24",
      "config": "[Interface]\nPrivateKey = 6Nu2payFq/fhwYolzFY1o3nXJZwq0+BkxGmoP10Uu3I=\nAddress = 192.168.32.3/24\nDNS = 8.8.8.8\n[Peer]\nPublicKey = njscYaHsusSQS77m2oVHN/kaooAaqGOTljOcYZicu38=\nAllowedIPs = 192.168.32.0/24\nEndpoint = 192.168.10.157:1002\nPersistentKeepalive = 20\n"
    }
  ]
}
```

---



