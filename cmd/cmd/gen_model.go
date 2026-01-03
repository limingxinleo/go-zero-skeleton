package cmd

import (
	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/spf13/cobra"
	"gorm.io/gen"
)

var genModelCmd = &cobra.Command{
	Use:   "gen:model {table_name}",
	Short: "Generate models for gorm",
	Long:  `Generate models for gorm`,
	Run: func(cmd *cobra.Command, args []string) {
		g := gen.NewGenerator(gen.Config{
			OutPath:      "./dal/query", // 生成的查询代码路径
			ModelPkgPath: "./dal/model", // 生成的模型路径

			// 生成查询时使用指针类型
			FieldNullable: true,
			// 生成字段的标签
			FieldCoverable: false,
			// 生成带签名的字段
			FieldSignable: false,
			// 生成索引标签
			FieldWithIndexTag: true,
			// 生成类型标签
			FieldWithTypeTag: true,
			// 生成单个模型文件
			Mode: gen.WithDefaultQuery | gen.WithQueryInterface,
		})

		ap := app.GetApplication()
		g.UseDB(ap.Gorm)

		if len(args) == 0 {
			g.ApplyBasic(g.GenerateAllTable()...)
		} else {
			for _, arg := range args {
				g.ApplyBasic(
					g.GenerateModel(arg),
				)
			}
		}

		g.Execute()
	},
}

func init() {
	rootCmd.AddCommand(genModelCmd)
}
