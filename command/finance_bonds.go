package command

import (
	"fmt"
	"github.com/jiangjiali/vault/sdk/helper/finance"
	"strconv"
	"strings"
	"time"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*FinanceBondsCommand)(nil)
var _ cli.CommandAutocomplete = (*FinanceBondsCommand)(nil)

type FinanceBondsCommand struct {
	*BaseCommand

	flagFunction   string
	flagSettlement string
	flagMaturity   string
	flagPrice      float64
	flagDiscount   float64
	flagRedemption float64
	flagBasis      int
}

func (c *FinanceBondsCommand) Synopsis() string {
	return "短期国库券"
}

func (c *FinanceBondsCommand) Help() string {
	helpText := `

使用: vault finance bonds [选项]

  债券计算公式
  都以100美元面值计算

  短期国库券的收益率：

      $ vault finance bonds -func=TBILLYIELD -settlement=2021/01/01 -maturity=2021/03/01 -price=80

  面值为￥100的短期国库券的发行价格：

      $ vault finance bonds -func=TBILLPRICE -settlement=2021/01/01 -maturity=2021/03/01 -discount=0.3

  短期国库券的等价收益率：

      $ vault finance bonds -func=TBILLEQ -settlement=2021/01/01 -maturity=2021/03/01 -discount=0.3

  有价证券的贴现率：

      $ vault finance bonds -func=DISC -settlement=2021/01/01 -maturity=2021/03/01 -price=100 -redemption=110 -basis=2

  折价发行的面值￥100的有价证券的价格：

      $ vault finance bonds -func=PRICEDISC -settlement=2021/01/01 -maturity=2021/03/01 -discount=0.3 -redemption=110 -basis=2

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *FinanceBondsCommand) Flags() *FlagSets {
	set := NewFlagSets()

	f := set.NewFlagSet("选项")

	f.StringVar(&StringVar{
		Name:       "func",
		Target:     &c.flagFunction,
		Default:    "TBILLYIELD",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "函数名字",
	})

	f.StringVar(&StringVar{
		Name:       "settlement",
		Target:     &c.flagSettlement,
		Default:    "2020/01/02",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "国库券的结算日。 即在发行日之后，国库券卖给购买者的日期。",
	})

	f.StringVar(&StringVar{
		Name:       "maturity",
		Target:     &c.flagMaturity,
		Default:    "2020/03/02",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "国库券的到期日。 到期日是国库券有效期截止时的日期。",
	})

	f.Float64Var(&Float64Var{
		Name:       "price",
		Target:     &c.flagPrice,
		Default:    100.0,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "面值￥100的国库券的价格。",
	})

	f.Float64Var(&Float64Var{
		Name:       "discount",
		Target:     &c.flagDiscount,
		Default:    0.0,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "国库券的贴现率。",
	})

	f.Float64Var(&Float64Var{
		Name:       "redemption",
		Target:     &c.flagRedemption,
		Default:    100.00,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "面值￥100的有价证券的清偿价值。",
	})

	f.IntVar(&IntVar{
		Name:       "basis",
		Target:     &c.flagBasis,
		Default:    0,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "要使用的日计数基准类型(0或省略-US(NASD)30/360、1-实际/实际、2-实际/360、3-实际/365、4-欧洲30/360)。",
	})

	return set
}

func (c *FinanceBondsCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *FinanceBondsCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *FinanceBondsCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	if len(args) > 0 {
		c.UI.Error(fmt.Sprintf("参数太多（应为 0 个，获得了 %d 个）", len(args)))
		return 1
	}

	if c.flagFunction == "" {
		c.UI.Error("函数名字是必须字段")
		return 1
	}

	// 转为时间戳
	settlement, _ := time.Parse("2006/01/02", c.flagSettlement)
	maturity, _ := time.Parse("2006/01/02", c.flagMaturity)

	switch c.flagFunction {
	case "TBILLYIELD":
		tbillYield, err := finance.TBillYield(settlement.Unix(), maturity.Unix(), c.flagPrice)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		reData := finance.StrconvFloat(tbillYield)
		c.UI.Output("短期国库券的收益率：" + reData)
	case "TBILLPRICE":
		tbillPrice, err := finance.TBillPrice(settlement.Unix(), maturity.Unix(), c.flagDiscount)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		reData := strconv.FormatFloat(tbillPrice, 'f', -1, 64)
		c.UI.Output("面值为￥100的短期国库券的发行价格：" + reData)
	case "TBILLEQ":
		tbillEq, err := finance.TBillEquivalentYield(settlement.Unix(), maturity.Unix(), c.flagDiscount)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		reData := finance.StrconvFloat(tbillEq)
		c.UI.Output("短期国库券的等价收益率：" + reData)
	case "DISC":
		discountRate := finance.DiscountRate(settlement.Unix(), maturity.Unix(), c.flagPrice, c.flagRedemption, c.flagBasis)
		reData := finance.StrconvFloat(discountRate)
		c.UI.Output("有价证券的贴现率：" + reData)
	case "PRICEDISC":
		priceDiscount := finance.PriceDiscount(settlement.Unix(), maturity.Unix(), c.flagDiscount, c.flagRedemption, c.flagBasis)
		reData := strconv.FormatFloat(priceDiscount, 'f', 2, 64)
		c.UI.Output("折价发行的面值￥100的有价证券的价格：" + reData)
	default:
		tbillyield, err := finance.TBillYield(settlement.Unix(), maturity.Unix(), c.flagPrice)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		reData := finance.StrconvFloat(tbillyield)
		c.UI.Output("短期国库券的收益率：" + reData)
	}

	return 0
}
