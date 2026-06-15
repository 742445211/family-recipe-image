package firdgemate

// CulinaryVision-YOLOv8n: 47 ingredient classes (HimanshuRay/CulinaryVision-YOLOv8n)
var culinaryLabels = []string{
	"almond", "apple", "asparagus", "avocado", "banana", "beans", "beet", "bell pepper",
	"blackberry", "blueberry", "broccoli", "brussels sprouts", "cabbage", "carrot",
	"cauliflower", "celery", "cherry", "corn", "cucumber", "egg", "eggplant", "garlic",
	"grape", "green bean", "green onion", "hot pepper", "kiwi", "lemon", "lettuce", "lime",
	"mandarin", "mushroom", "onion", "orange", "pattypan squash", "pea", "peach", "pear",
	"pineapple", "potato", "pumpkin", "radish", "raspberry", "strawberry", "tomato",
	"vegetable marrow", "watermelon",
}

var englishToChinese = map[string]string{
	"almond":           "杏仁",
	"apple":            "苹果",
	"asparagus":        "芦笋",
	"avocado":          "牛油果",
	"banana":           "香蕉",
	"beans":            "豆类",
	"beet":             "甜菜",
	"bell pepper":      "甜椒",
	"blackberry":       "黑莓",
	"blueberry":        "蓝莓",
	"broccoli":         "西兰花",
	"brussels sprouts": "抱子甘蓝",
	"cabbage":          "卷心菜",
	"carrot":           "胡萝卜",
	"cauliflower":      "花椰菜",
	"celery":           "芹菜",
	"cherry":           "樱桃",
	"corn":             "玉米",
	"cucumber":         "黄瓜",
	"egg":              "鸡蛋",
	"eggplant":         "茄子",
	"garlic":           "大蒜",
	"grape":            "葡萄",
	"green bean":       "四季豆",
	"green onion":      "葱",
	"hot pepper":       "辣椒",
	"kiwi":             "猕猴桃",
	"lemon":            "柠檬",
	"lettuce":          "生菜",
	"lime":             "青柠",
	"mandarin":         "橘子",
	"mushroom":         "蘑菇",
	"onion":            "洋葱",
	"orange":           "橙子",
	"pattypan squash":  "扁圆南瓜",
	"pea":              "豌豆",
	"peach":            "桃子",
	"pear":             "梨",
	"pineapple":        "菠萝",
	"potato":           "土豆",
	"pumpkin":          "南瓜",
	"radish":           "萝卜",
	"raspberry":        "树莓",
	"strawberry":       "草莓",
	"tomato":           "番茄",
	"vegetable marrow": "西葫芦",
	"watermelon":       "西瓜",
}

func ClassName(classID int, labels []string) string {
	if classID >= 0 && classID < len(labels) {
		return labels[classID]
	}
	return "unknown"
}

func ToChineseIngredient(english string) string {
	if cn, ok := englishToChinese[english]; ok {
		return cn
	}
	return english
}

func DefaultLabels(numClasses int) []string {
	if numClasses == len(culinaryLabels) || numClasses == 0 {
		return culinaryLabels
	}
	// fallback COCO subset for yolov8n (80 classes)
	return nil
}
