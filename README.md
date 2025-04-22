# Configuration Service

The configuration service provides an simple http endpoint which provides straight forward config map values as json files. It predefines no ingress so you have to declare one in the values.yaml file under ingress section.

Note: This service requires and service account in kubernetes with access to read config maps in the namespace.

# Use Case

The service provides for UIs or any other application a simple way to delivery config map content from the cluster.

# Functionality

Translates 1:1 a config map to rest api output. Example: 

Config Map:

```
data:
    example: value

```

Rest API:

``` 
{
    "example":"value"
}

```


