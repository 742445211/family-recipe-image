package firdgemate

// Roboflow ingredients-detection-yolov8-npkkb v3: 53 classes (family-v1 fine-tuned)
var ingredientLabels = []string{
	"1", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
	"2", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29",
	"3", "4", "5", "6", "7", "8", "9",
	"anchovy", "artichoke", "bell pepper", "broccoli", "cabbage", "carrot",
	"cauliflower", "chicken breast", "cucumber", "egg", "eggplant", "garlic",
	"green chilli pepper", "leek", "lemon", "lettuce", "mince", "onion",
	"parsley", "red meat", "tomato", "undefined", "white beans", "white button mushroom",
}

// CulinaryVision-YOLOv8n legacy: 47 classes (culinaryvision.onnx)
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
	"anchovy":               "鳀鱼",
	"artichoke":             "朝鲜蓟",
	"bell pepper":           "甜椒",
	"broccoli":              "西兰花",
	"cabbage":               "卷心菜",
	"carrot":                "胡萝卜",
	"cauliflower":           "花椰菜",
	"chicken breast":        "鸡胸肉",
	"cucumber":              "黄瓜",
	"egg":                   "鸡蛋",
	"eggplant":              "茄子",
	"garlic":                "大蒜",
	"green chilli pepper":   "青辣椒",
	"leek":                  "韭葱",
	"lemon":                 "柠檬",
	"lettuce":               "生菜",
	"mince":                 "肉末",
	"onion":                 "洋葱",
	"parsley":               "欧芹",
	"red meat":              "红肉",
	"tomato":                "番茄",
	"white beans":           "白豆",
	"white button mushroom": "白蘑菇",
	// legacy CulinaryVision
	"almond":           "杏仁",
	"apple":            "苹果",
	"asparagus":        "芦笋",
	"avocado":          "牛油果",
	"banana":           "香蕉",
	"beans":            "豆类",
	"beet":             "甜菜",
	"blackberry":       "黑莓",
	"blueberry":        "蓝莓",
	"brussels sprouts": "抱子甘蓝",
	"celery":           "芹菜",
	"cherry":           "樱桃",
	"corn":             "玉米",
	"grape":            "葡萄",
	"green bean":       "四季豆",
	"green onion":      "葱",
	"hot pepper":       "辣椒",
	"kiwi":             "猕猴桃",
	"lime":             "青柠",
	"mandarin":         "橘子",
	"mushroom":         "蘑菇",
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
	switch numClasses {
	case len(ingredientLabels):
		return ingredientLabels
	case len(culinaryLabels):
		return culinaryLabels
	default:
		if numClasses == 0 {
			return culinaryLabels
		}
		return nil
	}
}
