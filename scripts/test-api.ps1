<#
.SYNOPSIS
    Comprehensive API test script for OpenShield Manager
.DESCRIPTION
    Tests authentication, RBAC enforcement, organization CRUD,
    and auth middleware on existing endpoints.
#>

$BASE_URL = "https://localhost:9000"
$CURL = "curl.exe"

function Write-TestHeader {
    param([string]$Title)
    Write-Output "`n" ("="*60)
    Write-Output "  $Title"
    Write-Output ("="*60)
}

function Write-TestResult {
    param([string]$TestName, [bool]$Passed, [string]$Detail)
    $status = if ($Passed) { "PASS" } else { "FAIL" }
    Write-Output "[$status] $TestName"
    if ($Detail) { Write-Output "       $Detail" }
}

function Parse {
    param([string]$Json)
    try { return $Json | ConvertFrom-Json } catch { return $null }
}

function Login {
    param([string]$User, [string]$Pass)
    $result = & $CURL -sk -X POST "${BASE_URL}/api/auth/authenticate" -u "${User}:${Pass}"
    try { return ($result | ConvertFrom-Json).data.token } catch { return $null }
}

function CurlGet {
    param([string]$Path, [string]$Token)
    return & $CURL -sk -X GET "${BASE_URL}${Path}" -H "Authorization: Bearer ${Token}"
}

function CurlPostToken {
    param([string]$Path, [string]$Token, [string]$Json)
    return & $CURL -sk -X POST "${BASE_URL}${Path}" -H "Authorization: Bearer ${Token}" -H "Content-Type: application/json" -d $Json
}

function CurlDelete {
    param([string]$Path, [string]$Token)
    return & $CURL -sk -X DELETE "${BASE_URL}${Path}" -H "Authorization: Bearer ${Token}"
}

function CurlPut {
    param([string]$Path, [string]$Token, [string]$Json)
    return & $CURL -sk -X PUT "${BASE_URL}${Path}" -H "Authorization: Bearer ${Token}" -H "Content-Type: application/json" -d $Json
}

function Invoke-Api {
    param([string]$Method = "GET", [string]$Path, [string]$Token = $null, [object]$Body = $null)
    $args = @("-sk", "-X", $Method, "${BASE_URL}${Path}")
    if ($Token) { $args += @("-H", "Authorization: Bearer $Token") }
    if ($Body) {
        $tmpFile = [System.IO.Path]::GetTempFileName()
        Set-Content -Path $tmpFile -Value ($Body | ConvertTo-Json -Compress) -Encoding ASCII
        $args += @("-H", "Content-Type: application/json", "-d", "@$tmpFile")
    }
    $result = & $CURL @args
    if ($Body -and $tmpFile) { Remove-Item $tmpFile -ErrorAction SilentlyContinue }
    try { return $result | ConvertFrom-Json } catch { return $result }
}

# =====================================================================
# LOGIN
# =====================================================================
Write-TestHeader "SETUP: Login as super_admin"
$token = Login "admin@openshield.local" "admin"
Write-TestResult "Logged in as super_admin" ([bool]$token)

# =====================================================================
# 1. EXISTING ENDPOINTS - AUTH REQUIRED
# =====================================================================
Write-TestHeader "1. EXISTING ENDPOINTS - AUTH ENFORCEMENT"

Write-Output "`n--- 1.1 Existing endpoints WITHOUT auth (expect 401) ---"
$noAuth = & $CURL -sk -X GET "${BASE_URL}/api/agents/list" -w "%{http_code}"
Write-TestResult "GET /api/agents/list without auth" (($noAuth -match "Missing authorization header") -or ($noAuth -match "Authentication required"))

$noAuth2 = & $CURL -sk -X GET "${BASE_URL}/api/groups" -w "%{http_code}"
Write-TestResult "GET /api/groups without auth" (($noAuth2 -match "Missing authorization header") -or ($noAuth2 -match "Authentication required"))

$noAuth3 = & $CURL -sk -X GET "${BASE_URL}/api/jobs/list" -w "%{http_code}"
Write-TestResult "GET /api/jobs/list without auth" (($noAuth3 -match "Missing authorization header") -or ($noAuth3 -match "Authentication required"))

$noAuth4 = & $CURL -sk -X POST "${BASE_URL}/api/tools/execute" -H "Content-Type: application/json" -d '{"name":"test"}' -w "%{http_code}"
Write-TestResult "POST /api/tools/execute without auth" (($noAuth4 -match "Missing authorization header") -or ($noAuth4 -match "Authentication required"))

Write-Output "`n--- 1.2 Existing endpoints WITH auth (expect 200) ---"
$withAuth = CurlGet "/api/agents/list" $token
Write-TestResult "GET /api/agents/list with auth" (($withAuth | ConvertFrom-Json).error -eq 0 -or ($withAuth -match "data"))

$withAuth2 = CurlGet "/api/groups" $token
Write-TestResult "GET /api/groups with auth" (($withAuth2 | ConvertFrom-Json).error -eq 0 -or ($withAuth2 -match "data"))

$withAuth3 = CurlGet "/api/jobs/list" $token
Write-TestResult "GET /api/jobs/list with auth" (($withAuth3 | ConvertFrom-Json).error -eq 0 -or ($withAuth3 -match "data"))

$withAuth4 = CurlGet "/api/queries/list" $token
Write-TestResult "GET /api/queries/list with auth" (($withAuth4 | ConvertFrom-Json).error -eq 0 -or ($withAuth4 -match "data"))

$withAuth5 = CurlGet "/api/bulk-operations" $token
Write-TestResult "GET /api/bulk-operations with auth" (($withAuth5 | ConvertFrom-Json).error -eq 0 -or ($withAuth5 -match "data"))

# =====================================================================
# 2. RBAC - ROLE RESTRICTIONS
# =====================================================================
Write-TestHeader "2. RBAC - ROLE RESTRICTIONS"

Write-Output "`n--- 2.1 Register org_viewer user ---"
$viewer = Invoke-Api -Method POST -Path "/api/auth/register" -Token $token -Body @{ email = "viewer2@test.com"; password = "password123"; name = "Viewer Test"; role = "org_viewer" }
$viewerToken = Login "viewer2@test.com" "password123"
Write-TestResult "Viewer user created and logged in" ([bool]$viewerToken)

Write-Output "`n--- 2.2 viewer tries admin-level operations (expect 403) ---"
if ($viewerToken) {
    # viewer tries to delete a query (RoleAdmin required)
    $viewerDel = CurlDelete "/api/queries/some-id" $viewerToken
    $viewerDelResult = Parse $viewerDel
    Write-TestResult "viewer cannot delete query" ($viewerDelResult.error -eq "Insufficient permissions")

    # viewer tries to create a job (RoleOperator required)
    $viewerJob = CurlPostToken "/api/jobs/create" $viewerToken '{"name":"test","type":"script","target":"all"}'
    $viewerJobResult = Parse $viewerJob
    Write-TestResult "viewer cannot create job" ($viewerJobResult.error -eq "Insufficient permissions")

    # viewer tries to unregister agent (RoleAdmin required)
    $viewerUnreg = CurlPostToken "/api/agents/unregister" $viewerToken '{"id":"test-id"}'
    $viewerUnregResult = Parse $viewerUnreg
    Write-TestResult "viewer cannot unregister agent" ($viewerUnregResult.error -eq "Insufficient permissions")
}

Write-Output "`n--- 2.3 Register org_operator user ---"
$operator = Invoke-Api -Method POST -Path "/api/auth/register" -Token $token -Body @{ email = "operator2@test.com"; password = "password123"; name = "Operator Test"; role = "org_operator" }
$operatorToken = Login "operator2@test.com" "password123"
Write-TestResult "Operator user created and logged in" ([bool]$operatorToken)

Write-Output "`n--- 2.4 operator tries destructive operations (expect 403) ---"
if ($operatorToken) {
    # operator tries to delete a query (RoleAdmin required)
    $opDel = CurlDelete "/api/queries/some-id" $operatorToken
    $opDelResult = Parse $opDel
    Write-TestResult "operator cannot delete query" ($opDelResult.error -eq "Insufficient permissions")

    # operator tries to unregister agent (RoleAdmin required)
    $opUnreg = CurlPostToken "/api/agents/unregister" $operatorToken '{"id":"test-id"}'
    $opUnregResult = Parse $opUnreg
    Write-TestResult "operator cannot unregister agent" ($opUnregResult.error -eq "Insufficient permissions")

    # operator CAN create a job (RoleOperator required)
    $opJob = CurlPostToken "/api/jobs/create" $operatorToken '{"name":"test","type":"script","target":"all"}'
    $opJobResult = Parse $opJob
    # Should get past auth, might fail at validation/DB
    Write-TestResult "operator can attempt to create job" ($opJobResult.error -ne "Insufficient permissions" -and $opJobResult.error -ne "Authentication required")
}

# =====================================================================
# 3. AUTH ENDPOINTS STILL WORK
# =====================================================================
Write-TestHeader "3. AUTH ENDPOINTS REGRESSION CHECK"

Write-Output "`n--- 3.1 Login still works ---"
$reLogin = Login "admin@openshield.local" "admin"
Write-TestResult "Re-login still works" ([bool]$reLogin)

Write-Output "`n--- 3.2 Auth/me still works ---"
$meResult = CurlGet "/api/auth/me" $reLogin
$meObj = Parse $meResult
Write-TestResult "GET /api/auth/me works" ($meObj.error -eq 0 -and $meObj.data.email -eq "admin@openshield.local")

Write-Output "`n--- 3.3 List orgs still works ---"
$orgsResult = CurlGet "/api/organizations" $reLogin
$orgsObj = Parse $orgsResult
$defaultExists = $orgsObj.data | Where-Object { $_.slug -eq "default" }
Write-TestResult "List organizations works" ($defaultExists -ne $null)

Write-Output "`n" ("="*60)
Write-Output "  ALL TESTS COMPLETED"
Write-Output ("="*60)
