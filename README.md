# MySQL Dumper

*MySQL Dumper* is a tool for creating *filtered* and *manipulated* database dumps. It relies on the SQL native language, using WHERE clauses and complete SELECT statements with aliases.
The inspiration for this dumper comes from the [MySQL Super Dump](https://github.com/hgfischer/mysqlsuperdump)

Currently, it supports dumping tables and views.


## Features

* Filter by table name (config: `tables`)
* Filter by column name (config: `columns`)
* Filter by table size (config: `size`)
* Replace dumped column values with replacements (config: `replacements`)
* Disable data output of specific tables, dump only table data without definitions or completely ignore them (config: `filters`)
* Dumping to the file or directly to another database

## Config
 - `tables` - `key` is a table name and `value` is a filter (for example: `"users": "WHERE id = 1"`)
 - `columns` - `key` is a column name and `value` is a filter (for example: `"user_id": "WHERE user_id = 1"`). If any table has column that is specified here this filter will be used.
 - `size`
   - `gt` - value can be any that is acceptable by the `humanize.ParseBigBytes` ([link](https://github.com/dustin/go-humanize/blob/master/bigbytes.go))
   - `filters` - `key` can be table name or you can specify column name by placing the `*.` before column name and `value` is a filter (for examle: `"*.id": "ORDER BY id DESC LIMIT 30"`)
 - `replacements` - `key` consists of the table and column name and `value` is a replacement for the real value (for example: `"users.password": "MD5('123456')"`). So you can, for example, hide sensitive data if you are dumping the DB for the developers to use.
 - `filters` - `key` is table name and `value` is one of the following: `[onlydata, nodata, ignore]` (for example: `"logs": "nodata"`) You can also use `*` as a `key` by which the filter will be applied to all tables that do not match any other filter.

Values in the config are used in the following order:
1. table name is checked for existence in the `filters` part of the config. If table name exists and the value is:
    - `onlydata` - table definition will not be dumped, but the data will be
    - `nodata` - table definition will be dumped, but the data won't be
    - `ignore` - neither definition nor data will be dumped
2. table name is checked in the `tables` part of the config. If table name exists, the filter from the `value` is used for dumping the data.
3. columns from the table are checked for existence in the `columns` part of the config. The first column that exists in the table column list will be used for dumping the data.
4. table size will be checked against the value that you set as `gt` value (value that will be compared is `(information_schema.tables.data_length + information_schema.tables.index_length)`). The `key`'s under the `filters` are first checked for the table names. If table name exists then the `value` will be used for dumping the data. If table name does not exist, then column names will be checked. The first column that exists in the table column list will be used for dumping the data. If there is no matching filter, then the default filter `ORDER BY 1 DESC LIMIT 30` will be used.

## Configuration Example

```
{
    "tables": {
        "users": "WHERE id = 1",
        "carts": "WHERE user_id=1 AND item_id=2"
    },
    "columns": {
        "user_id": "WHERE user_id = 1",
        "cart_id": "WHERE cart_id = 3"
    },
    "size": {
        "gt": "10 MiB",
        "filters": {
            "*.id": "ORDER BY id DESC LIMIT 30",
            "*.created_at": "ORDER BY created_at DESC LIMIT 30",
            "items": "LIMIT 10"
        }
    },
    "replacements": {
        "users.password": "MD5('123456')"
    },
    "filters": {
        "table1": "onlydata",
        "table2": "nodata",
        "table3": "ignore"
    }
}
```

## Usage
 - [dump to a file](examples/file-dumper/main.go)
 - [dump to Stdout](examples/stdout-dumper/main.go)
 - [dump to another database](examples/db-dumper/main.go) *IMPORTANT*: `multiStatements` must be enabled