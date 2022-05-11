package validator

import (
	"testing"
)

func submit1() []byte {
	msg := `{"params": ["stage.s9x211", "783647bc", "8372000000000000", "61e6f66c", "0a3f74a7"], "id": 16801, "method": "mining.submit"}`
	bmsg := []byte(msg)
	return bmsg
}

func submit2() []byte {
	msg := ``
	bmsg := []byte(msg)
	return bmsg
}

func notify1() []byte {
	msg := `{"id":6190,"jsonrpc":"2.0","method":"mining.notify","params":["616c4a28","17c2c0507d5b4f32aa1ca39d82b83f16dfbf75d000093fd60000000000000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2803cbf90a00046cf6e6610c","0a746974616e2f6a74677261737369650affffffff024aa07e26000000001976a9143f1ad6ada38343e55cc4c332657e30a71c86b66188ac0000000000000000266a24aa21a9ed5617dd59c856a6ae1c00b6df9c4ed26727616d4d7b59f3eaacd16d810ff6dd3400000000",[ "e03d3ffb98db39658948c3e7e612b18d60a863b981b76d4fe17717d77818e4a3", "a3bc1f07702d7d17411fa3a7d686efc4ac6e45d7b3f3d5f6091194afcf5ce9ab", "5c60807c2e560c0ca85edc301136a4f3f442dc738fc2896cebd0088d7273d7ab", "62248c534cc4cf906f63ce71b3662bb986430c563225d7106ab1f14791ca6d90", "f957f25e5fac2ee6e1de76b3272ad5ddc751414ef2bc1d67fdcb50484eea9be7", "300c97cc0ce4179ef67dd0d974fd7f70bcb7f4d71a8b675f7f2955c780d94ecf", "243796dcef2ee48f1a5132c42abb9ab11ae47ddf87f77797390ef0cf0cdd3a3b", "5e26cbaa9f32657aac5be852e9c560e429f80019fda2da39fee69685086325f2", "df44bb117da9c9c79919d78a16255715fd4c62aa03e6b67c4667907ac897e15a", "94a8c79e827851ebb6f043e393db73111e754a7b0b55fb0c83c6cbc6235f1603", "1db100d99f37b95d429df28a11f0bef873dc63c8510787e2c9be199662236f06", "680aa7cd5a2007a9f2ae2a76ed2fb2e3231c8b6cd3ccaeea951a089a91912223" ],"20000000","170b8c8b","61e6f66c",false]}`
	bmsg := []byte(msg)
	return bmsg
}

func notify2() []byte {
	msg := `{"id":5896,"jsonrpc":"2.0","method":"mining.notify","params":["783647bc","17c2c0507d5b4f32aa1ca39d82b83f16dfbf75d000093fd60000000000000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2803cbf90a00046cf6e6610c","0a746974616e2f6a74677261737369650affffffff024aa07e26000000001976a9143f1ad6ada38343e55cc4c332657e30a71c86b66188ac0000000000000000266a24aa21a9ed5617dd59c856a6ae1c00b6df9c4ed26727616d4d7b59f3eaacd16d810ff6dd3400000000",[ "e03d3ffb98db39658948c3e7e612b18d60a863b981b76d4fe17717d77818e4a3", "a3bc1f07702d7d17411fa3a7d686efc4ac6e45d7b3f3d5f6091194afcf5ce9ab", "5c60807c2e560c0ca85edc301136a4f3f442dc738fc2896cebd0088d7273d7ab", "62248c534cc4cf906f63ce71b3662bb986430c563225d7106ab1f14791ca6d90", "f957f25e5fac2ee6e1de76b3272ad5ddc751414ef2bc1d67fdcb50484eea9be7", "300c97cc0ce4179ef67dd0d974fd7f70bcb7f4d71a8b675f7f2955c780d94ecf", "243796dcef2ee48f1a5132c42abb9ab11ae47ddf87f77797390ef0cf0cdd3a3b", "5e26cbaa9f32657aac5be852e9c560e429f80019fda2da39fee69685086325f2", "df44bb117da9c9c79919d78a16255715fd4c62aa03e6b67c4667907ac897e15a", "94a8c79e827851ebb6f043e393db73111e754a7b0b55fb0c83c6cbc6235f1603", "1db100d99f37b95d429df28a11f0bef873dc63c8510787e2c9be199662236f06", "680aa7cd5a2007a9f2ae2a76ed2fb2e3231c8b6cd3ccaeea951a089a91912223" ],"20000000","170b8c8b","61e6f66c",false]}`
	bmsg := []byte(msg)
	return bmsg
}

func TestNotifyJSON(t *testing.T) {
	m := notify1()
	_, err := convertJSONToNotify(m)

	if err != nil {
		t.Errorf("error when parsing notify message: %+v", err)
	}
}

func TestSubmitJSON(t *testing.T) {
	m := submit1()
	_, err := convertJSONToSubmit(m)

	if err != nil {
		t.Errorf("error when parsing submit message: %+v", err)
	}
}
