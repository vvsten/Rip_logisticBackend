# AuthApi

All URIs are relative to *http://localhost:8083*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**loginPost**](#loginpost) | **POST** /login | User login|
|[**logoutPost**](#logoutpost) | **POST** /logout | User logout|
|[**refreshPost**](#refreshpost) | **POST** /refresh | Refresh access token|
|[**signUpPost**](#signuppost) | **POST** /sign_up | Register new user|

# **loginPost**
> ServiceAuthResponse loginPost(request)

Authenticate user with login and password

### Example

```typescript
import {
    AuthApi,
    Configuration,
    ServiceLoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AuthApi(configuration);

let request: ServiceLoginRequest; //Login credentials

const { status, data } = await apiInstance.loginPost(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **ServiceLoginRequest**| Login credentials | |


### Return type

**ServiceAuthResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Login successful |  -  |
|**400** | Invalid request |  -  |
|**401** | Invalid credentials |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **logoutPost**
> { [key: string]: string; } logoutPost()

Logout user and invalidate tokens

### Example

```typescript
import {
    AuthApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AuthApi(configuration);

const { status, data } = await apiInstance.logoutPost();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**{ [key: string]: string; }**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Logout successful |  -  |
|**401** | Unauthorized |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **refreshPost**
> ServiceAuthResponse refreshPost(request)

Refresh access token using refresh token

### Example

```typescript
import {
    AuthApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AuthApi(configuration);

let request: { [key: string]: string; }; //Refresh token

const { status, data } = await apiInstance.refreshPost(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **{ [key: string]: string; }**| Refresh token | |


### Return type

**ServiceAuthResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tokens refreshed successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Invalid refresh token |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **signUpPost**
> ServiceAuthResponse signUpPost(request)

Register a new user with login, email, password and other details

### Example

```typescript
import {
    AuthApi,
    Configuration,
    ServiceRegisterRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AuthApi(configuration);

let request: ServiceRegisterRequest; //Registration data

const { status, data } = await apiInstance.signUpPost(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **ServiceRegisterRequest**| Registration data | |


### Return type

**ServiceAuthResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | User registered successfully |  -  |
|**400** | Invalid request |  -  |
|**409** | User already exists |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

