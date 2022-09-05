-- depends_on: {{ ref('TRANS01_DATA') }}
{{ config(tags=['PROJECT', 'DATA']) }}
{{  clone_table('TRANS01_DATA', "DATA") }}