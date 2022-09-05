{{ config(tags=['PROJECT', 'DATA']) }}

SELECT 
    "V":id::string AS ID,
    "V":name::string AS NAME,
FROM 
    {{ source('SOURCE_NAME', 'TABLE_NAME')}}
