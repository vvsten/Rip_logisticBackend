# LogisticRequestsApi

All URIs are relative to *http://localhost:8083*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**apiLogisticRequestsGet**](#apilogisticrequestsget) | **GET** /api/logistic-requests | Get logistic requests list|
|[**apiLogisticRequestsIdCompletePut**](#apilogisticrequestsidcompleteput) | **PUT** /api/logistic-requests/{id}/complete | Complete or reject logistic request|
|[**apiLogisticRequestsPost**](#apilogisticrequestspost) | **POST** /api/logistic-requests | Create cargo logistic request|

# **apiLogisticRequestsGet**
> { [key: string]: any; } apiLogisticRequestsGet()

Get logistic requests list with filtering by status and date range

### Example

```typescript
import {
    LogisticRequestsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new LogisticRequestsApi(configuration);

let status: string; //Logistic request status filter (optional) (default to undefined)
let dateFrom: string; //Date from (YYYY-MM-DD) (optional) (default to undefined)
let dateTo: string; //Date to (YYYY-MM-DD) (optional) (default to undefined)

const { status, data } = await apiInstance.apiLogisticRequestsGet(
    status,
    dateFrom,
    dateTo
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **status** | [**string**] | Logistic request status filter | (optional) defaults to undefined|
| **dateFrom** | [**string**] | Date from (YYYY-MM-DD) | (optional) defaults to undefined|
| **dateTo** | [**string**] | Date to (YYYY-MM-DD) | (optional) defaults to undefined|


### Return type

**{ [key: string]: any; }**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Logistic requests retrieved successfully |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiLogisticRequestsIdCompletePut**
> { [key: string]: string; } apiLogisticRequestsIdCompletePut(request)

Complete or reject logistic request by moderator

### Example

```typescript
import {
    LogisticRequestsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new LogisticRequestsApi(configuration);

let id: number; //Logistic request ID (default to undefined)
let request: { [key: string]: string; }; //Logistic request status (completed/rejected)

const { status, data } = await apiInstance.apiLogisticRequestsIdCompletePut(
    id,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **{ [key: string]: string; }**| Logistic request status (completed/rejected) | |
| **id** | [**number**] | Logistic request ID | defaults to undefined|


### Return type

**{ [key: string]: string; }**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Logistic request completed successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiLogisticRequestsPost**
> { [key: string]: any; } apiLogisticRequestsPost(request)

Create a new cargo transportation logistic request

### Example

```typescript
import {
    LogisticRequestsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new LogisticRequestsApi(configuration);

let request: { [key: string]: any; }; //Logistic request data with services

const { status, data } = await apiInstance.apiLogisticRequestsPost(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **{ [key: string]: any; }**| Logistic request data with services | |


### Return type

**{ [key: string]: any; }**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Logistic request submitted successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Unauthorized |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

