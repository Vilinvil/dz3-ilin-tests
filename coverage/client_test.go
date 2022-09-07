package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const (
	ErrorLimitBelowZero  = "limit must be > 0"
	ErrorOffsetBelowZero = "offset must be > 0"
	ErroeBadAccessToken  = "bad AccessToken"
	ErrorwrongDataSet    = "SearchServer fatal error. Body: {\"Error\":\"couldn't read file wrongDataSet. Error is: open wrongDataSet: no such file or directory\"}"
	ErrorTimeout         = "timeout for limit=2&offset=0&order_by=1&order_field=Id&query="
	ErrorWrongUrl        = "unknown error Get \"wrongUrl?limit=2&offset=0&order_by=1&order_field=Id&query=\": unsupported protocol scheme \"\""
)

type TestCaseParam struct {
	Url      string
	Response string
}

type TestCaseErrors struct {
	Url        string
	Response   string
	StatusCode int
}

type TestCasePatchDataSet struct {
	Url              string
	TestPatchDataSet string
	Response         SearchErrorResponse
	StatusCode       int
}

type TestCaseClient struct {
	Request     SearchRequest
	Result      *SearchResponse
	ResponseErr error
}

func TestSearchServerParam(t *testing.T) {
	cases := []TestCaseParam{
		// Cases order_field=Name and different order_by
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Name&order_by=1",
			Response: `[{"ID":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"},{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"}]`,
		},
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Name&order_by=-1",
			Response: `[{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"},{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"}]`,
		},
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Name&order_by=0",
			Response: `[{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"}]`,
		},

		// Cases order_field=ID and different order_by
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Id&order_by=1",
			Response: `[{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"}]`,
		},
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Id&order_by=-1",
			Response: `[{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"}]`,
		},
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Id&order_by=0",
			Response: `[{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"}]`,
		},

		// Cases order_field=Age and different order_by
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Age&order_by=1",
			Response: `[{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"}]`,
		},
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Age&order_by=-1",
			Response: `[{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"},{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"}]`,
		},
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=1&query=Nulla&order_field=Age&order_by=0",
			Response: `[{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"}]`,
		},

		// Case of max records(26)
		{
			Url:      "https://127.0.0.1:8080/?limit=26&offset=0&query=&order_field=Id&order_by=1",
			Response: `[{"ID":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"},{"ID":1,"Name":"Hilda Mayer","Age":21,"About":"Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n","Gender":"female"},{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":3,"Name":"Everett Dillard","Age":27,"About":"Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n","Gender":"male"},{"ID":4,"Name":"Owen Lynn","Age":30,"About":"Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n","Gender":"male"},{"ID":5,"Name":"Beulah Stark","Age":30,"About":"Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n","Gender":"female"},{"ID":6,"Name":"Jennings Mays","Age":39,"About":"Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n","Gender":"male"},{"ID":7,"Name":"Leann Travis","Age":34,"About":"Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.\n","Gender":"female"},{"ID":8,"Name":"Glenn Jordan","Age":29,"About":"Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n","Gender":"male"},{"ID":9,"Name":"Rose Carney","Age":36,"About":"Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n","Gender":"female"},{"ID":10,"Name":"Henderson Maxwell","Age":30,"About":"Ex et excepteur anim in eiusmod. Cupidatat sunt aliquip exercitation velit minim aliqua ad ipsum cillum dolor do sit dolore cillum. Exercitation eu in ex qui voluptate fugiat amet.\n","Gender":"male"},{"ID":11,"Name":"Gilmore Guerra","Age":32,"About":"Labore consectetur do sit et mollit non incididunt. Amet aute voluptate enim et sit Lorem elit. Fugiat proident ullamco ullamco sint pariatur deserunt eu nulla consectetur culpa eiusmod. Veniam irure et deserunt consectetur incididunt ad ipsum sint. Consectetur voluptate adipisicing aute fugiat aliquip culpa qui nisi ut ex esse ex. Sint et anim aliqua pariatur.\n","Gender":"male"},{"ID":12,"Name":"Cruz Guerrero","Age":36,"About":"Sunt enim ad fugiat minim id esse proident laborum magna magna. Velit anim aliqua nulla laborum consequat veniam reprehenderit enim fugiat ipsum mollit nisi. Nisi do reprehenderit aute sint sit culpa id Lorem proident id tempor. Irure ut ipsum sit non quis aliqua in voluptate magna. Ipsum non aliquip quis incididunt incididunt aute sint. Minim dolor in mollit aute duis consectetur.\n","Gender":"male"},{"ID":13,"Name":"Whitley Davidson","Age":40,"About":"Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n","Gender":"male"},{"ID":14,"Name":"Nicholson Newman","Age":23,"About":"Tempor minim reprehenderit dolore et ad. Irure id fugiat incididunt do amet veniam ex consequat. Quis ad ipsum excepteur eiusmod mollit nulla amet velit quis duis ut irure.\n","Gender":"male"},{"ID":15,"Name":"Allison Valdez","Age":21,"About":"Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n","Gender":"male"},{"ID":16,"Name":"Annie Osborn","Age":35,"About":"Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n","Gender":"female"},{"ID":17,"Name":"Dillard Mccoy","Age":36,"About":"Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n","Gender":"male"},{"ID":18,"Name":"Terrell Hall","Age":27,"About":"Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n","Gender":"male"},{"ID":19,"Name":"Bell Bauer","Age":26,"About":"Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n","Gender":"male"},{"ID":20,"Name":"Lowery York","Age":27,"About":"Dolor enim sit id dolore enim sint nostrud deserunt. Occaecat minim enim veniam proident mollit Lorem irure ex. Adipisicing pariatur adipisicing aliqua amet proident velit. Magna commodo culpa sit id.\n","Gender":"male"},{"ID":21,"Name":"Johns Whitney","Age":26,"About":"Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n","Gender":"male"},{"ID":22,"Name":"Beth Wynn","Age":31,"About":"Proident non nisi dolore id non. Aliquip ex anim cupidatat dolore amet veniam tempor non adipisicing. Aliqua adipisicing eu esse quis reprehenderit est irure cillum duis dolor ex. Laborum do aute commodo amet. Fugiat aute in excepteur ut aliqua sint fugiat do nostrud voluptate duis do deserunt. Elit esse ipsum duis ipsum.\n","Gender":"female"},{"ID":23,"Name":"Gates Spencer","Age":21,"About":"Dolore magna magna commodo irure. Proident culpa nisi veniam excepteur sunt qui et laborum tempor. Qui proident Lorem commodo dolore ipsum.\n","Gender":"male"},{"ID":24,"Name":"Gonzalez Anderson","Age":33,"About":"Quis consequat incididunt in ex deserunt minim aliqua ea duis. Culpa nisi excepteur sint est fugiat cupidatat nulla magna do id dolore laboris. Aute cillum eiusmod do amet dolore labore commodo do pariatur sit id. Do irure eiusmod reprehenderit non in duis sunt ex. Labore commodo labore pariatur ex minim qui sit elit.\n","Gender":"male"},{"ID":25,"Name":"Katheryn Jacobs","Age":32,"About":"Magna excepteur anim amet id consequat tempor dolor sunt id enim ipsum ea est ex. In do ea sint qui in minim mollit anim est et minim dolore velit laborum. Officia commodo duis ut proident laboris fugiat commodo do ex duis consequat exercitation. Ad et excepteur ex ea exercitation id fugiat exercitation amet proident adipisicing laboris id deserunt. Commodo proident laborum elit ex aliqua labore culpa ullamco occaecat voluptate voluptate laboris deserunt magna.\n","Gender":"female"}]`,
		},

		// Case of empty order_field
		{
			Url:      "https://127.0.0.1:8080/?limit=4&offset=0&query=&order_field=&order_by=1",
			Response: `[{"ID":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"},{"ID":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},{"ID":3,"Name":"Everett Dillard","Age":27,"About":"Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n","Gender":"male"},{"ID":1,"Name":"Hilda Mayer","Age":21,"About":"Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n","Gender":"female"}]`,
		},

		// Case offset > limit
		{
			Url:      "https://127.0.0.1:8080/?limit=2&offset=10&query=&order_field=Id&order_by=1",
			Response: `[{"ID":10,"Name":"Henderson Maxwell","Age":30,"About":"Ex et excepteur anim in eiusmod. Cupidatat sunt aliquip exercitation velit minim aliqua ad ipsum cillum dolor do sit dolore cillum. Exercitation eu in ex qui voluptate fugiat amet.\n","Gender":"male"},{"ID":11,"Name":"Gilmore Guerra","Age":32,"About":"Labore consectetur do sit et mollit non incididunt. Amet aute voluptate enim et sit Lorem elit. Fugiat proident ullamco ullamco sint pariatur deserunt eu nulla consectetur culpa eiusmod. Veniam irure et deserunt consectetur incididunt ad ipsum sint. Consectetur voluptate adipisicing aute fugiat aliquip culpa qui nisi ut ex esse ex. Sint et anim aliqua pariatur.\n","Gender":"male"}]`,
		},
	}
	for caseNum, item := range cases {
		req := httptest.NewRequest("GET", item.Url, nil)
		req.Header.Add("AccessToken", "2a54a886a8bbcc309ae4ffa75241cd6d")
		w := httptest.NewRecorder()
		SearchServer(w, req)
		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%d] failed to read body: %v", caseNum, err)
		}

		bodyStr := string(body)
		if bodyStr != item.Response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, bodyStr, item.Response)
		}
		resp.Body.Close()
	}
}

func TestSearchServerErrors(t *testing.T) {
	cases := []TestCaseErrors{
		{
			Url:        "https://127.0.0.1:8080/?limit=tr&offset=24&query=&order_field=Age&order_by=1",
			Response:   ErrorBadLimit,
			StatusCode: 400,
		},
		{
			Url:        "https://127.0.0.1:8080/?limit=4&offset=rf&query=&order_field=Age&order_by=1",
			Response:   ErrorBadOffset,
			StatusCode: 400,
		},
		{
			Url:        "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=AgeIncorrect&order_by=1",
			Response:   ErrorBadOrderField,
			StatusCode: 400,
		},
		{
			Url:        "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=Age&order_by=OrderByInvalid",
			Response:   ErrorBadOrderBy,
			StatusCode: 400,
		},
		{
			Url:        "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=Name&order_by=10",
			Response:   ErrorBadOrderBy,
			StatusCode: 400,
		},
		{
			Url:        "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=Age&order_by=10",
			Response:   ErrorBadOrderBy,
			StatusCode: 400,
		},
		{
			Url:        "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=Id&order_by=10",
			Response:   ErrorBadOrderBy,
			StatusCode: 400,
		},
	}
	for caseNum, item := range cases {
		req := httptest.NewRequest("GET", item.Url, nil)
		req.Header.Add("AccessToken", "2a54a886a8bbcc309ae4ffa75241cd6d")
		w := httptest.NewRecorder()
		SearchServer(w, req)
		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%d] failed to read body: %v", caseNum, err)
		}

		var structErr SearchErrorResponse
		err = json.Unmarshal(body, &structErr)
		if err != nil {
			fmt.Printf("%+v\n", structErr)
			t.Errorf("couldn't json.Unmarshall %v. Error is %v", string(body), err)
		}

		if resp.StatusCode != item.StatusCode {
			t.Errorf("[%d] wrong StatusCode: got %+v, expected %+v",
				caseNum, resp.StatusCode, item.StatusCode)
		}

		if structErr.Error != item.Response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, structErr.Error, item.Response)
		}
		resp.Body.Close()
	}
}

func TestSearchServerPatchDataSet(t *testing.T) {
	cases := []TestCasePatchDataSet{
		{
			Url:              "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=Age&order_by=1",
			TestPatchDataSet: "dataSetForTests/dataSetNoXml.xml",
			Response:         SearchErrorResponse{Error: "couldn't parse file dataSetForTests/dataSetNoXml.xml. Error is: XML syntax error on line 37: attribute name without = in element"},
			StatusCode:       500,
		},
		{
			Url:              "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=Age&order_by=1",
			TestPatchDataSet: "dataSetForTests/dataSetWrongId.xml",
			Response:         SearchErrorResponse{Error: "in dataSetForTests/dataSetWrongId.xml incorrect id. Error is: strconv.Atoi: parsing \"ghg\": invalid syntax"},
			StatusCode:       500,
		},
		{
			Url:              "https://127.0.0.1:8080/?limit=26&offset=24&query=&order_field=Age&order_by=1",
			TestPatchDataSet: "dataSetForTests/dataSetWrongAge.xml",
			Response:         SearchErrorResponse{Error: "in dataSetForTests/dataSetWrongAge.xml incorrect age Error is: strconv.Atoi: parsing \"Twenty\": invalid syntax"},
			StatusCode:       500,
		},
	}
	defer func() {
		PatchDataSet = "dataset.xml"
	}()
	for caseNum, item := range cases {
		PatchDataSet = item.TestPatchDataSet
		req := httptest.NewRequest("GET", item.Url, nil)
		req.Header.Add("AccessToken", "2a54a886a8bbcc309ae4ffa75241cd6d")
		w := httptest.NewRecorder()
		SearchServer(w, req)
		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%d] failed to read body: %v", caseNum, err)
		}

		if resp.StatusCode != item.StatusCode {
			t.Errorf("[%d] wrong StatusCode: got %+v, expected %+v",
				caseNum, resp.StatusCode, item.StatusCode)
		}

		data := &SearchErrorResponse{}
		err = json.Unmarshal(body, data)

		if data.Error != item.Response.Error {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, data.Error, item.Response.Error)
		}
		resp.Body.Close()
	}
}

func TestClient(t *testing.T) {
	cases := []TestCaseClient{
		{
			Request: SearchRequest{Limit: 1,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    1},

			Result: &SearchResponse{Users: []User{{ID: 0,
				Name:   "Boyd Wolf",
				Age:    22,
				About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
				Gender: "male"}},
				NextPage: true},

			ResponseErr: nil,
		},
		{
			Request: SearchRequest{Limit: -1,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    1},

			Result:      nil,
			ResponseErr: fmt.Errorf(ErrorLimitBelowZero),
		},
		{
			Request: SearchRequest{Limit: 50,
				Offset:     0,
				Query:      "This text don't find",
				OrderField: "Id",
				OrderBy:    1},

			Result:      &SearchResponse{Users: nil, NextPage: false},
			ResponseErr: nil,
		},
		{
			Request: SearchRequest{Limit: 2,
				Offset:     -1,
				Query:      "",
				OrderField: "Id",
				OrderBy:    1},

			Result:      nil,
			ResponseErr: fmt.Errorf(ErrorOffsetBelowZero),
		},
		{
			Request: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "BadField",
				OrderBy:    1},
			Result:      nil,
			ResponseErr: fmt.Errorf("OrderFeld %s invalid", "BadField"),
		},
		{
			Request: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    987654},
			Result:      nil,
			ResponseErr: fmt.Errorf("unknown bad request error: %s", ErrorBadOrderBy),
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	testClient := SearchClient{AccessToken: "2a54a886a8bbcc309ae4ffa75241cd6d", URL: ts.URL}

	for caseNum, item := range cases {
		result, err := testClient.FindUsers(item.Request)
		if !reflect.DeepEqual(err, item.ResponseErr) {
			t.Errorf("[%d] got unexpected error: %#v, expected: %#v", caseNum, err, item.ResponseErr)
		}

		if item.ResponseErr != nil && err == nil {
			t.Errorf("[%d] got: %v expected error: %#v", caseNum, err, item.ResponseErr)
		}

		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, got: %#v, expected: %#v,", caseNum, result, item.Result)
		}
	}
	ts.Close()
}

func TestClientSpecificError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	testClient := SearchClient{AccessToken: "wrongToken", URL: ts.URL}
	req := SearchRequest{Limit: 1,
		Offset:     0,
		Query:      "",
		OrderField: "Id",
		OrderBy:    1}

	result, err := testClient.FindUsers(req)
	if err.Error() != ErroeBadAccessToken {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	testClient.AccessToken = "2a54a886a8bbcc309ae4ffa75241cd6d"

	client.Timeout = time.Microsecond
	result, err = testClient.FindUsers(req)
	if err.Error() != ErrorTimeout {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	client.Timeout = time.Second

	PatchDataSet = "wrongDataSet"
	result, err = testClient.FindUsers(req)
	if err.Error() != ErrorwrongDataSet {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	PatchDataSet = "dataset.xml"

	testClient.URL = "wrongUrl"
	result, err = testClient.FindUsers(req)
	if err.Error() != ErrorWrongUrl {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	ts.Close()
}

/*
	go test -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html

*/
