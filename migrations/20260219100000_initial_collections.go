package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// 1. Create 'studies' collection
		studies := core.NewBaseCollection("studies")
		studies.Fields.Add(
			&core.TextField{Name: "title"},
			&core.TextField{Name: "id_no"},
			&core.TextField{Name: "abstract"},
			&core.TextField{Name: "time_period"},
			&core.TextField{Name: "nation"},
			&core.TextField{Name: "universe"},
			&core.TextField{Name: "author"},
			&core.TextField{Name: "author_affiliation"},
			&core.TextField{Name: "producer"},
			&core.TextField{Name: "producer_affiliation"},
			&core.TextField{Name: "holdings_uri"},
			&core.TextField{Name: "holdings_description"},
			&core.TextField{Name: "analysis_unit"},
			&core.TextField{Name: "data_kind"},
			&core.JSONField{Name: "topic_classifications", MaxSize: 65536}, // 64 KB
			&core.JSONField{Name: "keywords", MaxSize: 65536},              // 64 KB
		)
		if err := app.Save(studies); err != nil {
			return err
		}

		// 2. Create 'variable_groups' collection
		groups := core.NewBaseCollection("variable_groups")
		groups.Fields.Add(
			&core.RelationField{
				Name:          "study",
				CollectionId:  studies.Id,
				Required:      true,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.TextField{Name: "ddi_id"},
			&core.TextField{Name: "name"},
			&core.TextField{Name: "concept", Required: true},
			&core.TextField{Name: "description"},
			&core.TextField{Name: "type"},
			&core.NumberField{Name: "order"},
		)
		if err := app.Save(groups); err != nil {
			return err
		}

		// 3. Create 'variables' collection
		variables := core.NewBaseCollection("variables")
		variables.Fields.Add(
			&core.RelationField{
				Name:          "study",
				CollectionId:  studies.Id,
				Required:      true,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.RelationField{
				Name:         "group",
				CollectionId: groups.Id,
				MaxSelect:    1,
			},
			&core.TextField{Name: "ddi_id"},
			&core.TextField{Name: "name"},
			&core.TextField{Name: "concept", Required: true},
			&core.TextField{Name: "question"},
			&core.TextField{Name: "prequestion_text"},
			&core.TextField{Name: "ivu_instructions"},
			&core.TextField{Name: "interval"},
			&core.TextField{Name: "var_format_type"},
			&core.TextField{Name: "answer_type"},
			&core.BoolField{Name: "has_other"},
			&core.BoolField{Name: "has_long_list"},
			&core.TextField{Name: "long_list_standard"},
			&core.JSONField{Name: "categories", MaxSize: 524288}, // 512 KB
			&core.NumberField{Name: "order"},
		)
		if err := app.Save(variables); err != nil {
			return err
		}

		// 4. Set read rules on collections (write stays admin-only)
		publicRule := ""
		authRule := "@request.auth.id != ''"

		// Studies: public read (browsed by unauthenticated visitors)
		if c, _ := app.FindCollectionByNameOrId("studies"); c != nil {
			c.ListRule = &publicRule
			c.ViewRule = &publicRule
			app.Save(c)
		}
		// Variables and variable groups: auth-only via PocketBase CRUD
		// (public access is through the custom /api/questions endpoints instead)
		for _, name := range []string{"variable_groups", "variables"} {
			c, _ := app.FindCollectionByNameOrId(name)
			if c != nil {
				c.ListRule = &authRule
				c.ViewRule = &authRule
				app.Save(c)
			}
		}

		return nil
	}, nil)
}
