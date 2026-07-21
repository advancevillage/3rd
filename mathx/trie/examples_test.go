package trie

import (
	"strings"
	"testing"
)

func Test_Example1(t *testing.T) {
	filter, err := NewFilter()
	if err != nil {
		t.Fatalf("new filter: %v", err)
	}
	filter.AddWord("一个东西")
	filter.AddWord("一些")
	filter.AddWord("个")
	filter.AddWord("有一")
	filter.AddWord("一个")
	filter.AddWord("有一个东西")
	filter.AddWord("有一个")
	filter.AddWord("东西")
	filter.AddWord("个东")
	filter.AddWord("哈哈")

	t.Logf(">>> FindAll <<<")
	t.Logf("result: %v", filter.FindAll("有一个东西东西哈哈"))

	t.Logf(">>> Replace <<<")
	t.Logf("result: %v", filter.Replace("hello", '*'))
	t.Logf("result: %v", filter.Replace("骚", '*'))
	t.Logf("result: %v", filter.Replace("一", '*'))
	t.Logf("result: %v", filter.Replace("一个", '*'))
	t.Logf("result: %v", filter.Replace("一个东", '*'))
	t.Logf("result: %v", filter.Replace("一个东西", '*'))
	t.Logf("result: %v", filter.Replace("一个东西啊", '*'))
	t.Logf("result: %v", filter.Replace("有一个东西啊", '*'))
	t.Logf("result: %v", filter.Replace("有一个东啊", '*'))
	t.Logf("result: %v", filter.Replace("有一个啊", '*'))
	t.Logf("result: %v", filter.Replace("有一个", '*'))
	t.Logf("result: %v", filter.Replace("有一", '*'))

	t.Logf(">>> Validate <<<")
	t.Log(filter.Validate("一"))
	t.Log(filter.Validate("一个"))
	t.Log(filter.Validate("一个东"))
	t.Log(filter.Validate("一个东西"))
	t.Log(filter.Validate("一个东西啊"))
	t.Log(filter.Validate("有一个东西啊"))
	t.Log(filter.Validate("有一个东啊"))
	t.Log(filter.Validate("有一个啊"))
	t.Log(filter.Validate("有一个"))
	t.Log(filter.Validate("有一"))

	t.Logf(">>> Filter <<<")
	t.Logf("result: %v", filter.Filter("一"))
	t.Logf("result: %v", filter.Filter("一个"))
	t.Logf("result: %v", filter.Filter("一个东"))
	t.Logf("result: %v", filter.Filter("一个东西"))
	t.Logf("result: %v", filter.Filter("一个东西啊"))
	t.Logf("result: %v", filter.Filter("有一个东西啊"))
	t.Logf("result: %v", filter.Filter("有一个东啊"))
	t.Logf("result: %v", filter.Filter("有一个啊"))
	t.Logf("result: %v", filter.Filter("有一个"))
	t.Logf("result: %v", filter.Filter("有一"))
}

func Test_Example2(t *testing.T) {
	filter, err := NewFilter()
	if err != nil {
		t.Fatalf("new filter: %v", err)
	}
	words := "交友,约会,Soul,陌陌,探探,Uki,漂流瓶,伊对,Blued,LOFTER,牵手,cp速配,相遇,社交,同城,聊天,秘友,附近密聊,他趣,阿姨,美女,小伙," +
		"大叔,好友,儿童泳衣,哺乳衣,招商,加盟,占卜,总裁,秘书,穿越,甜妻,娇妻,老公,老婆,男朋友,女朋友,新郎,新娘,大结局,开局,游戏,重生,王爷,王妃," +
		"相公,虐文,渣男,修仙,离婚,重活,前世,女帝,虐哭,战神,宠文,爽文,纳妾,前夫,女婿,永生,手游,玩法,闯关,几关,塔防,模拟经营,养成,开心盒子,乐园," +
		"作战,打仗,三国,武器,射击,枪,装备,法师,斗地主,麻将,捕鱼,对决,格斗,机甲,奔爱,保险,小说,同心,剑舞九天,灵猫,缘分聊,聊缘,斗鱼,泰康保,对象," +
		"相亲,离异,珍爱,寂寞,脱单,小姐姐,配对,匹配,语音,情感,陌生人,可约,邀请,无聊,性感,胸部,备孕,登录奖励,零钱包,现金,点击收下,额度,提现,网贷," +
		"网商银行,无限金币,暴击,阴阳鱼,跑酷,超级加倍,入账,仙侠,天姬变,妖灵,江湖,酒,派派,TapTap,话费,大曲,退款,阅读,恋爱,学长,学姐,美甲小屋,剑," +
		"无限火力,漫画,陈酿,五粮液,茅台,股票,首充,七猫,火线少女,西游,免费看书吧,SSR,快看,知道了,金融,祛斑,祛痘,医美,淡斑,脱发,新氧,单身,校园模拟器," +
		"原神,立即领取,微众银行,私信,缘分,脱毛,洋葱学园,美女主播,虎牙,yy,YY,茎部,男人,内衣,保健,云漫读,云短剧,起点,晃一晃,本故事纯属虚构,炒股," +
		"九游,投资,签到"
	filter.AddWord(strings.Split(words, ",")...)

	textList := []string{
		"这是一款下载app",
		"这是一款下载交友app",
		"这是一款下载约会app",
		"这是一款下载交友和约会的app",
		"这是一款下载交会的app",
	}
	t.Logf(">>> Validate <<<")
	for _, text := range textList {
		t.Logf("result: %v", filter.FindAll(text))
		t.Log(filter.Validate(text))
		t.Log("\n")
	}
}
