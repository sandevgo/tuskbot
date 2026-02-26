# Telegram Channel Specification

TuskBot utilizes the Telegram Bot API as its primary communication interface. Access is restricted via a strict ownership verification mechanism.

## Access Control Configuration

The system implements a mandatory whitelist to prevent unauthorized resource consumption and data access.

### Ownership Verification

| Variable | Type | Description |
| :--- | :--- | :--- |
| `TUSK_TELEGRAM_OWNER_ID` | `int64` | The unique numeric identifier of the authorized Telegram user. |

### Operational Constraints
- **Inbound Filtering**: All messages originating from identifiers not matching `TUSK_TELEGRAM_OWNER_ID` are discarded at the transport layer.
- **Privacy**: The bot does not respond to unauthorized users, ensuring the agent remains private to the operator.

## Provisioning

1. Obtain a Bot Token from [@BotFather](https://t.me/botfather).
2. Retrieve your numeric User ID (e.g., via [@userinfobot](https://t.me/userinfobot)).
3. Assign these values to `TUSK_TELEGRAM_TOKEN` and `TUSK_TELEGRAM_OWNER_ID` respectively.
