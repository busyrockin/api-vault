from fastmcp import FastMCP
from fastmcp.exceptions import ToolError

from vault import Vault, VaultError
from approvals import Approvals

mcp = FastMCP("api-vault")
vault = Vault()
approvals = Approvals()


@mcp.tool()
def get_credential(name: str) -> str:
    """Retrieve a decrypted API key by name."""
    try:
        key = vault.get(name)
    except VaultError as e:
        raise ToolError(str(e))
    approvals.log_access(name)
    return key


@mcp.tool()
def list_credentials() -> list[dict]:
    """List all stored credential names, types, and creation dates."""
    try:
        return vault.list()
    except VaultError as e:
        raise ToolError(str(e))


if __name__ == "__main__":
    mcp.run()
