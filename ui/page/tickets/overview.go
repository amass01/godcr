package tickets

import (
	"fmt"

	"gioui.org/layout"

	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/page/components"
	"github.com/planetdecred/godcr/ui/values"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

const OverviewPageID = "Tickets"

type Page struct {
	*load.Load

	ticketPageContainer *layout.List
	ticketsLive         *layout.List

	purchaseTicket decredmaterial.Button

	ticketPrice  string
	totalRewards string

	autoPurchaseEnabled *decredmaterial.Switch
	toTickets           decredmaterial.TextAndIconButton

	stakingOverview *dcrlibwallet.StakingOverview
	liveTickets     []*transactionItem
}

func NewTicketPage(l *load.Load) *Page {
	pg := &Page{
		Load: l,

		ticketsLive:         &layout.List{Axis: layout.Horizontal},
		ticketPageContainer: &layout.List{Axis: layout.Vertical},
		purchaseTicket:      l.Theme.Button("Purchase"),

		autoPurchaseEnabled: l.Theme.Switch(),
		toTickets:           l.Theme.TextAndIconButton("See All", l.Icons.NavigationArrowForward),
	}

	pg.toTickets.Color = l.Theme.Color.Primary
	pg.toTickets.BackgroundColor = l.Theme.Color.Surface

	pg.stakingOverview = new(dcrlibwallet.StakingOverview)
	return pg
}

func (pg *Page) ID() string {
	return OverviewPageID
}

func (pg *Page) OnResume() {

	pg.loadPageData()

	go pg.WL.GetVSPList()
	// TODO: automatic ticket purchase functionality
	pg.autoPurchaseEnabled.Disabled()
}

func (pg *Page) loadPageData() {
	go func() {
		ticketPrice, err := pg.WL.MultiWallet.TicketPrice()
		if err != nil {
			pg.Toast.NotifyError(err.Error())
		} else {
			pg.ticketPrice = dcrutil.Amount(ticketPrice.TicketPrice).String()
			pg.RefreshWindow()
		}
	}()

	go func() {
		totalRewards, err := pg.WL.MultiWallet.TotalStakingRewards()
		if err != nil {
			pg.Toast.NotifyError(err.Error())
		} else {
			pg.totalRewards = dcrutil.Amount(totalRewards).String()
			pg.RefreshWindow()
		}
	}()

	go func() {
		overview, err := pg.WL.MultiWallet.StakingOverview()
		if err != nil {
			pg.Toast.NotifyError(err.Error())
		} else {
			pg.stakingOverview = overview
			pg.RefreshWindow()
		}
	}()

	go func() {
		mw := pg.WL.MultiWallet
		tickets, err := allLiveTickets(mw)
		if err != nil {
			pg.Toast.NotifyError(err.Error())
			return
		}

		txItems, err := ticketsToTransactionItems(pg.Load, tickets, true, func(filter int32) bool {
			switch filter {
			case dcrlibwallet.TxFilterUnmined:
				fallthrough
			case dcrlibwallet.TxFilterImmature:
				fallthrough
			case dcrlibwallet.TxFilterLive:
				return true
			}

			return false
		})
		if err != nil {
			pg.Toast.NotifyError(err.Error())
			return
		}

		pg.liveTickets = txItems
		pg.RefreshWindow()
	}()
}

func (pg *Page) Layout(gtx layout.Context) layout.Dimensions {
	return components.UniformPadding(gtx, func(gtx layout.Context) layout.Dimensions {
		sections := []func(gtx C) D{
			func(ctx layout.Context) layout.Dimensions {
				return pg.ticketPriceSection(gtx)
			},
			func(ctx layout.Context) layout.Dimensions {
				return pg.ticketsLiveSection(gtx)
			},
			func(ctx layout.Context) layout.Dimensions {
				return pg.stakingRecordSection(gtx)
			},
		}
		return pg.ticketPageContainer.Layout(gtx, len(sections), func(gtx C, i int) D {
			return sections[i](gtx)
		})
	})
}

func (pg *Page) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	return layout.Inset{
		Bottom: values.MarginPadding8,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return pg.Theme.Card().Layout(gtx, func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.UniformInset(values.MarginPadding16).Layout(gtx, body)
		})
	})
}

func (pg *Page) titleRow(gtx layout.Context, leftWidget, rightWidget func(C) D) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(leftWidget),
		layout.Rigid(rightWidget),
	)
}

func (pg *Page) ticketPriceSection(gtx layout.Context) layout.Dimensions {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding11,
				}.Layout(gtx, func(gtx C) D {
					// leftWg := func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							title := pg.Theme.Label(values.TextSize14, "Ticket Price")
							title.Color = pg.Theme.Color.Gray2
							return title.Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{
								Left:  values.MarginPadding8,
								Right: values.MarginPadding4,
							}.Layout(gtx, func(gtx C) D {
								ic := pg.Icons.TimerIcon
								return ic.Layout12dp(gtx)
							})
						}),
						layout.Rigid(func(gtx C) D {
							secs, _ := pg.WL.MultiWallet.NextTicketPriceRemaining()
							txt := pg.Theme.Label(values.TextSize14, nextTicketRemaining(int(secs)))
							txt.Color = pg.Theme.Color.Gray2
							return txt.Layout(gtx)
						}),
					)
					// }
					//TODO: auto ticket purchase.
					// return pg.titleRow(gtx, leftWg, pg.autoPurchaseEnabled.Layout)-
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding8,
				}.Layout(gtx, func(gtx C) D {
					ic := pg.Icons.TicketPurchasedIcon
					return layout.Center.Layout(gtx, ic.Layout48dp)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding16,
				}.Layout(gtx, func(gtx C) D {
					return layout.Center.Layout(gtx, func(gtx C) D {
						return components.LayoutBalanceSize(gtx, pg.Load, pg.ticketPrice, values.TextSize28)
					})
				})
			}),
			layout.Rigid(pg.purchaseTicket.Layout),
		)
	})
}

func (pg *Page) ticketsLiveSection(gtx layout.Context) layout.Dimensions {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: values.MarginPadding14}.Layout(gtx, func(gtx C) D {
					title := pg.Theme.Label(values.TextSize14, "Live Tickets")
					title.Color = pg.Theme.Color.Gray
					return pg.titleRow(gtx, title.Layout, func(gtx C) D {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							pg.stakingCountIcon(pg.Icons.TicketUnminedIcon, pg.stakingOverview.Unmined),
							pg.stakingCountIcon(pg.Icons.TicketImmatureIcon, pg.stakingOverview.Immature),
							pg.stakingCountIcon(pg.Icons.TicketLiveIcon, pg.stakingOverview.Live),
							layout.Rigid(pg.toTickets.Layout),
						)
					})
				})
			}),
			layout.Rigid(func(gtx C) D {
				return pg.ticketsLive.Layout(gtx, len(pg.liveTickets), func(gtx C, index int) D {
					return layout.Inset{Right: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
						return ticketCard(gtx, pg.Load, pg.liveTickets[index], true)
					})
				})
			}),
		)
	})
}

func (pg *Page) stakingCountIcon(icon *decredmaterial.Image, count int) layout.FlexChild {
	return layout.Rigid(func(gtx C) D {
		if count == 0 {
			return D{}
		}
		return layout.Inset{Right: values.MarginPadding14}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return icon.Layout16dp(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
						label := pg.Theme.Label(values.TextSize14, fmt.Sprintf("%d", count))
						label.Color = pg.Theme.Color.DeepBlue
						return label.Layout(gtx)
					})
				}),
			)
		})
	})
}

func (pg *Page) stakingRecordSection(gtx C) D {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding14,
				}.Layout(gtx, func(gtx C) D {
					title := pg.Theme.Label(values.TextSize14, "Staking Record")
					title.Color = pg.Theme.Color.Gray2
					return pg.titleRow(gtx, title.Layout, func(gtx C) D { return D{} })
				})
			}),
			layout.Rigid(func(gtx C) D {
				wdgs := []layout.Widget{
					pg.stakingRecordIconCount(pg.Icons.TicketUnminedIcon, pg.stakingOverview.Unmined, "Unmined"),
					pg.stakingRecordIconCount(pg.Icons.TicketImmatureIcon, pg.stakingOverview.Immature, "Immature"),
					pg.stakingRecordIconCount(pg.Icons.TicketLiveIcon, pg.stakingOverview.Live, "Live"),
					pg.stakingRecordIconCount(pg.Icons.TicketVotedIcon, pg.stakingOverview.Voted, "Voted"),
					pg.stakingRecordIconCount(pg.Icons.TicketExpiredIcon, pg.stakingOverview.Expired, "Expired"),
					pg.stakingRecordIconCount(pg.Icons.TicketRevokedIcon, pg.stakingOverview.Revoked, "Revoked"),
				}

				return decredmaterial.GridWrap{
					Axis:      layout.Horizontal,
					Alignment: layout.End,
				}.Layout(gtx, len(wdgs), func(gtx C, i int) D {
					return wdgs[i](gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return decredmaterial.LinearLayout{
					Width:       decredmaterial.MatchParent,
					Height:      decredmaterial.WrapContent,
					Background:  pg.Theme.Color.Success2,
					Padding:     layout.Inset{Top: values.MarginPadding16, Bottom: values.MarginPadding16},
					Border:      decredmaterial.Border{Radius: decredmaterial.Radius(8)},
					Direction:   layout.Center,
					Orientation: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Bottom: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
							txt := pg.Theme.Label(values.TextSize14, "Rewards Earned")
							txt.Color = pg.Theme.Color.Success
							return txt.Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								ic := pg.Icons.StakeyIcon
								return ic.Layout24dp(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								return components.LayoutBalance(gtx, pg.Load, pg.totalRewards)
							}),
						)
					}),
				)
			}),
		)
	})
}

func (pg *Page) stakingRecordIconCount(icon *decredmaterial.Image, count int, status string) layout.Widget {
	return func(gtx C) D {
		return layout.Inset{Bottom: values.MarginPadding16, Right: values.MarginPadding40}.Layout(gtx, func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return icon.Layout24dp(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								label := pg.Theme.Label(values.TextSize16, fmt.Sprintf("%d", count))
								label.Color = pg.Theme.Color.DeepBlue
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								txt := pg.Theme.Label(values.TextSize12, status)
								txt.Color = pg.Theme.Color.Gray
								return txt.Layout(gtx)
							}),
						)
					})
				}),
			)
		})
	}
}

func (pg *Page) Handle() {
	if pg.purchaseTicket.Clicked() {
		newTicketPurchaseModal(pg.Load).
			TicketPurchased(func() {
				fmt.Println("Overview ticket pruchsased")
				pg.loadPageData()
			}).Show()
	}

	if pg.toTickets.Button.Clicked() {
		pg.ChangeFragment(newListPage(pg.Load))
	}
}

func (pg *Page) OnClose() {}
