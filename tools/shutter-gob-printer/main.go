package main

import (
	"fmt"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/app"
)

// func main() {
// 	dkgResultHex := "7b7f03010106526573756c7401ff800001070103456f6e010600010a4e756d4b65797065727301060001095468726573686f6c6401060001064b6579706572010600010e5365637265744b6579536861726501ff820001095075626c69634b657901ff8400010f5075626c69634b657953686172657301ff880000000aff81050102ff8a0000000aff83050102ff8c0000002cff870201011d5b5d2a736863727970746f2e456f6e5075626c69634b6579536861726501ff880001ff8600000aff85050102ff8e000000fe0337ff800104010701040221027357a7abefc52a2bf1fff35238a0b436aa4d49a7e58fd555823753f90c55c73d016096e157e0535e4398d616676a75d90f87cb080dd4472bd52a21d091bf3a3e376fdf19d194a7a8d9fb457d46742e7917010fb534d7b9ac40fdd9a6c44246e3aa5bfd64c8c5430da8ebaaaa47e5b6accdc419c223c64e63e9defd4ae5c397b76fc90107608cce993feed43f24a7d4cb8c6c98633bc55ab83eab75ba7429c32bd1630e80ef757fdf43ecaa884f716060a0e18e5c9417ad9b5edee15a3dcad835cbebbdc25cf649698b14b244117f09cb9de2b49f01951d03bfcc751f9ff7cc9c5598004fb4608cc5f4d14c5cc57f3c1656b4fbefc16d07c345dac23e3964e1df30c2c038e7bf049606f0eaf29d048436021a8c4074db14f5ac47b5d4ff4e4b78acf3b34b066335a7f50e5fc81ecd48aa6b46cb1b0b5c38b97a5bf325277349d98c1002df3e2d60a12a47d8efab79ba8c0633dc5880f889689ddaaa10ef922159001a20228a31ef72d013e480f99ccd548b734b47e2c600187b0b9a039d43335f883b703106da9e02ff0f99027e52619e1262820f7a014b590d44710b500e478f78d7edcc2fc1a360b1721bd12866657ba30356a0cf433a42f052122df5fb9191b9f16dffebbd21c6fd620ab36a727104bc668f307026427808fc5f91a166dc378752fd307a187fa6b38754e73209445185682af332cf258426c94037cf78b673d3e45eceb90e9d1460a8a78c16ad8355159a4497eab2df4efc978df2cd5363edaacc698eaac65c84d7eeaf5e86eea7df08bd14ecc585e6d8b117757a1018004f8d1a5fa9c1f6a46a2d11e0ec82c9c5a5dadd68b25b950878ddaa862c366392b39d64a1816582aeae7560829b29ec59187d1fcebb63a515fd32822465150cb74cd8d6d035c250c31e045a850793469eba13d7e7e1039715da196004ba5c6610c85213306aff231d77e37951f689c9ffaabea9f79cd929697a30a6a17f2b5def90b4569d2a899875c4b2df60a8c45b98c8b79fb9b7b76403b6d9cc3c0b5f749dd177849947d837f28d3c241fa4e35d83c6adf5ca9ce630eff10a8280118ef8bc90c42c80fda5dd94d96ae80fa65a551745893c3c413e5a0c665c3110f67bee296b89e64b78d6a79486e7964300"
// 	dkgResultBytes := common.FromHex(dkgResultHex)
// 	dkgResult, err := shdb.DecodePureDKGResult(dkgResultBytes)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("%+v\n", dkgResult)
// }

func main() {
	a, err := app.LoadShutterAppFromFile("/Users/ulo/t/shutter2.gob")
	if err != nil {
		panic(err)
	}
	printShutterApp(&a)

	for i, config := range a.Configs {
		n := a.CountCheckedInKeypers(config.Keypers)
		k := app.NumRequiredTransitionValidators(config)
		fmt.Printf("Config %d has %d checked-in keypers, required: %d\n", i, n, k)
	}
}

func printShutterApp(app *app.ShutterApp) {
	fmt.Println("ShutterApp Data:")

	fmt.Println("Configs:")
	for i, config := range app.Configs {
		fmt.Printf("  Config %d: %+v\n", i, config)
	}

	fmt.Println("DKGMap:")
	for eon, instance := range app.DKGMap {
		fmt.Printf("  Eon %d: %+v\n", eon, instance)
	}

	fmt.Printf("ConfigVoting: %+v\n", app.ConfigVoting)

	fmt.Printf("Gobpath: %s\n", app.Gobpath)
	fmt.Printf("LastSaved: %s\n", app.LastSaved)
	fmt.Printf("LastBlockHeight: %d\n", app.LastBlockHeight)

	fmt.Println("Identities:")
	for address, pubkey := range app.Identities {
		fmt.Printf("  Address: %v, Pubkey: %+v\n", address, pubkey)
	}

	fmt.Println("BlocksSeen:")
	for address, blocks := range app.BlocksSeen {
		fmt.Printf("  Address: %v, BlocksSeen: %d\n", address, blocks)
	}

	fmt.Printf("Validators: %+v\n", app.Validators)
	fmt.Printf("EONCounter: %d\n", app.EONCounter)
	fmt.Printf("DevMode: %t\n", app.DevMode)
	// fmt.Printf("CheckTxState: %+v\n", app.CheckTxState)
	// fmt.Printf("NonceTracker: %+v\n", app.NonceTracker)
	fmt.Printf("ChainID: %s\n", app.ChainID)
}
