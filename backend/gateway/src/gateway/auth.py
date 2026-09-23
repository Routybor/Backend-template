from fastapi import HTTPException, Request
from starlette.concurrency import run_in_threadpool

from common.auth import TokenValidator


async def require_user(request: Request) -> str:
    auth_header = request.headers.get("authorization")
    if not auth_header:
        raise HTTPException(401, {"error": "missing authorization header"})

    parts = auth_header.split(" ", 1)
    if len(parts) != 2 or parts[0].lower() != "bearer":
        raise HTTPException(401, {"error": "invalid authorization header format"})

    validator: TokenValidator = request.app.state.validator
    try:
        sub = await run_in_threadpool(validator.validate, parts[1])
    except Exception:
        raise HTTPException(401, {"error": "invalid token"}) from None

    return sub
