# Environment Variables Reference

# =============================================================================
# NAMESPACE
# =============================================================================

K8S_NAMESPACE
    Description: Kubernetes namespace for all resources
    Default: microservices
    Type: string

# =============================================================================
# KEYCLOAK
# =============================================================================

KEYCLOAK_IMAGE
    Description: Keycloak container image
    Default: quay.io/keycloak/keycloak:26.5.0
    Type: string (image reference)

KEYCLOAK_ADMIN
    Description: Keycloak admin username
    Default: admin
    Type: string

KEYCLOAK_ADMIN_PASSWORD
    Description: Keycloak admin password
    Default: admin
    Type: string (sensitive)

KEYCLOAK_DB
    Description: Keycloak database type
    Default: dev-file
    Options: dev-file, postgres, mariadb, mysql
    Type: string

KEYCLOAK_LOG_LEVEL
    Description: Keycloak logging level
    Default: INFO
    Options: DEBUG, INFO, WARN, ERROR
    Type: string

KEYCLOAK_REALM
    Description: Keycloak realm name
    Default: microservices
    Type: string

KEYCLOAK_CLIENT_ID
    Description: Keycloak client ID for gateway authentication
    Default: gateway
    Type: string

KEYCLOAK_CLIENT_SECRET
    Description: Keycloak client secret
    Default: gateway-secret
    Type: string (sensitive)

KEYCLOAK_TESTUSER_PASSWORD
    Description: Password for test user in realm
    Default: testuser
    Type: string (sensitive)

# =============================================================================
# CORE SERVICE
# =============================================================================

CORE_SERVICE_IMAGE
    Description: Core service container image
    Default: ultimatetemplate/core-service:latest
    Type: string (image reference)

CORE_SERVICE_REPLICAS
    Description: Number of core-service pod replicas
    Default: 2
    Type: integer

CORE_SERVICE_PORT
    Description: Core service container port
    Default: 8081
    Type: integer

CORE_SERVICE_GRPC_PORT
    Description: Core service gRPC port
    Default: 9091
    Type: integer

# =============================================================================
# GATEWAY
# =============================================================================

GATEWAY_IMAGE
    Description: Gateway container image
    Default: ultimatetemplate/gateway:latest
    Type: string (image reference)

GATEWAY_REPLICAS
    Description: Number of gateway pod replicas
    Default: 2
    Type: integer

GATEWAY_PORT
    Description: Gateway container port
    Default: 8080
    Type: integer

# =============================================================================
# STORAGE
# =============================================================================

STORAGE_CLASS
    Description: Kubernetes storage class for PVCs
    Default: standard
    Type: string

KEYCLOAK_DATA_SIZE
    Description: Size of Keycloak data PVC
    Default: 1Gi
    Type: string (quantity)
