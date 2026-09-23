import logging

import jwt
from jwt import PyJWKClient

logger = logging.getLogger(__name__)


class TokenValidator:
    def __init__(self, keycloak_url: str, realm: str) -> None:
        self._jwks = PyJWKClient(
            f"{keycloak_url}/realms/{realm}/protocol/openid-connect/certs"
        )
        self._realm = realm

    def validate(self, token: str) -> str:
        key = self._jwks.get_signing_key_from_jwt(token).key
        payload = jwt.decode(
            token,
            key,
            algorithms=["RS256"],
            options={"verify_iss": False, "verify_aud": False},
        )

        iss = payload.get("iss", "")
        if not iss.endswith(f"/realms/{self._realm}"):
            raise jwt.InvalidTokenError("invalid issuer")

        sub = payload.get("sub")
        if not sub:
            raise jwt.InvalidTokenError("missing subject claim")

        return sub
