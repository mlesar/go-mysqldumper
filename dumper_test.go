package mysqldumper

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestViewReordering(t *testing.T) {
	Convey("Should reorder views", t, func() {
		dumper := New(nil, nil, nil)

		viewsIn := &Tables{
			Table{Name: "view2", Type: "VIEW", Definition: "CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`%` SQL SECURITY DEFINER VIEW `view2` AS SELECT id FROM `view1`;"},
			Table{Name: "view1", Type: "VIEW", Definition: "CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`%` SQL SECURITY DEFINER VIEW `view1` AS SELECT id FROM `table1`;"},
			Table{Name: "view3", Type: "VIEW", Definition: "CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`%` SQL SECURITY DEFINER VIEW `view3` AS SELECT id FROM `table2` JOIN `view2` ON (`view2`.`id`=`table2`.`id`);"},
		}

		viewsOut := &Tables{
			Table{Name: "view1", Type: "VIEW", Definition: "CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`%` SQL SECURITY DEFINER VIEW `view1` AS SELECT id FROM `table1`;"},
			Table{Name: "view2", Type: "VIEW", Definition: "CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`%` SQL SECURITY DEFINER VIEW `view2` AS SELECT id FROM `view1`;"},
			Table{Name: "view3", Type: "VIEW", Definition: "CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`%` SQL SECURITY DEFINER VIEW `view3` AS SELECT id FROM `table2` JOIN `view2` ON (`view2`.`id`=`table2`.`id`);"},
		}

		dumper.reorderViews(viewsIn)

		So(viewsIn, ShouldResemble, viewsOut)
	})
}
