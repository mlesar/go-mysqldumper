package mysqldumper

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var sampleConfig = `
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
		"table3": "ignore",
		"*": "onlydata"
    }
}
`

func TestDumFilter(t *testing.T) {
	Convey("Parse config file", t, func() {
		_, err := ParseConfig([]byte(sampleConfig))
		So(err, ShouldBeNil)
	})

	Convey("Check config parameters", t, func() {
		dumper, err := ParseConfig([]byte(sampleConfig))
		So(err, ShouldBeNil)

		filter := dumper.GetDumpFilter(Table{Name: "users"})
		So(filter, ShouldEqual, "WHERE id = 1")

		filter = dumper.GetDumpFilter(Table{Name: "carts"})
		So(filter, ShouldEqual, "WHERE user_id=1 AND item_id=2")

		filter = dumper.GetDumpFilter(Table{Name: "items", Columns: &Columns{"id", "cart_id", "name", "value"}})
		So(filter, ShouldEqual, "WHERE cart_id = 3")

		size := Size(10485760) // ==10 MiB
		filter = dumper.GetDumpFilter(Table{Name: "logs", Size: &size})
		So(filter, ShouldEqual, "")

		size = Size(10485760 + 1) // >10 MiB
		filter = dumper.GetDumpFilter(Table{Name: "logs", Size: &size})
		So(filter, ShouldEqual, "ORDER BY 1 DESC LIMIT 30")

		filter = dumper.GetDumpFilter(Table{Name: "items", Size: &size})
		So(filter, ShouldEqual, "LIMIT 10")

		filter = dumper.GetDumpFilter(Table{Name: "table1"})
		So(filter, ShouldEqual, "")

		filter = dumper.GetDumpFilter(Table{Name: "table2"})
		So(filter, ShouldEqual, "LIMIT 0")

		filter = dumper.GetDumpFilter(Table{Name: "table3"})
		So(filter, ShouldEqual, "LIMIT 0")

		filter = dumper.GetDumpFilter(Table{Name: "table4"})
		So(filter, ShouldEqual, "")

		canDump := dumper.CanDumpDefinition("table1")
		So(canDump, ShouldBeFalse)

		canDump = dumper.CanDumpDefinition("table2")
		So(canDump, ShouldBeTrue)

		canDump = dumper.CanDumpDefinition("table3")
		So(canDump, ShouldBeFalse)

		canDump = dumper.CanDumpDefinition("table4")
		So(canDump, ShouldBeFalse)

		columns := dumper.GetDumpColumns(Table{Name: "users", Columns: &Columns{"id", "name", "password"}})
		So(columns, ShouldEqual, "`id`,`name`,MD5('123456') AS `password`")
	})

	Convey("Check empty config", t, func() {
		dumper, err := ParseConfig(nil)
		So(err, ShouldBeNil)

		filter := dumper.GetDumpFilter(Table{Name: "users"})
		So(filter, ShouldEqual, "")

		filter = dumper.GetDumpFilter(Table{Name: "carts"})
		So(filter, ShouldEqual, "")

		filter = dumper.GetDumpFilter(Table{Name: "items", Columns: &Columns{"id", "cart_id", "name", "value"}})
		So(filter, ShouldEqual, "")

		size := Size(10485760) // ==10 MiB
		filter = dumper.GetDumpFilter(Table{Name: "logs", Size: &size})
		So(filter, ShouldEqual, "")

		size = Size(10485760 + 1) // >10 MiB
		filter = dumper.GetDumpFilter(Table{Name: "logs", Size: &size})
		So(filter, ShouldEqual, "")

		filter = dumper.GetDumpFilter(Table{Name: "table1"})
		So(filter, ShouldEqual, "")

		filter = dumper.GetDumpFilter(Table{Name: "table2"})
		So(filter, ShouldEqual, "")

		filter = dumper.GetDumpFilter(Table{Name: "table3"})
		So(filter, ShouldEqual, "")
	})
}
