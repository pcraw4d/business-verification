package word_segmentation

// loadBusinessDictionary loads a comprehensive dictionary of business-related words
// Includes common business terms, industry-specific terms, and common English words
func loadBusinessDictionary() map[string]bool {
	dict := make(map[string]bool)

	// Common business terms
	businessTerms := []string{
		"shop", "store", "market", "retail", "commerce", "business", "company", "corp", "inc", "llc", "ltd",
		"group", "enterprise", "services", "solutions", "systems", "tech", "digital", "online", "web", "net",
		"cyber", "smart", "global", "world", "international", "national", "local", "regional",
		"center", "hub", "zone", "place", "spot", "mall", "mart", "outlet", "boutique", "emporium",
		"restaurant", "cafe", "bar", "pub", "bistro", "diner", "tavern", "eatery", "kitchen", "dining",
		"food", "beverage", "wine", "beer", "coffee", "tea", "bakery", "catering", "delivery", "takeout",
		"hotel", "resort", "inn", "lodge", "motel", "hostel", "accommodation", "hospitality",
		"bank", "finance", "financial", "credit", "loan", "mortgage", "insurance", "investment", "trading",
		"health", "medical", "clinic", "hospital", "pharmacy", "dental", "wellness", "fitness", "gym",
		"education", "school", "university", "college", "academy", "institute", "training", "learning",
		"real", "estate", "property", "rental", "leasing", "construction", "building", "development",
		"transport", "transportation", "logistics", "shipping", "delivery", "freight", "courier",
		"manufacturing", "production", "factory", "warehouse", "distribution", "supply", "chain",
		"technology", "software", "hardware", "app", "application", "platform", "system", "network",
		"consulting", "advisory", "professional", "legal", "accounting", "audit", "tax",
		"marketing", "advertising", "media", "communication", "public", "relations", "pr",
		"design", "creative", "art", "graphic", "web", "digital", "multimedia",
		"entertainment", "music", "film", "video", "gaming", "sports", "recreation",
		"agriculture", "farming", "fishery", "mining", "energy", "utilities", "power",
		"wholesale", "trade", "import", "export", "international", "global",
	}

	// Industry-specific terms
	industryTerms := []string{
		"wine", "wines", "winery", "vineyard", "vintner", "sommelier", "tasting", "cellar", "bottle", "vintage",
		"grape", "grapes", "grapevine", "oenology", "alcohol", "spirits", "liquor", "brewery", "distillery",
		"retailer", "merchant", "dealer", "vendor", "seller", "reseller", "distributor", "wholesaler",
		"ecommerce", "e-commerce", "online", "digital", "internet", "cyber", "virtual",
		"startup", "start", "up", "venture", "capital", "innovation", "disrupt", "disruption",
		"fintech", "fin", "tech", "paytech", "insurtech", "proptech", "edtech", "healthtech",
		"saas", "software", "platform", "cloud", "api", "sdk", "framework", "library",
		"ai", "artificial", "intelligence", "machine", "learning", "deep", "neural", "network",
		"blockchain", "crypto", "bitcoin", "ethereum", "nft", "defi", "web3",
		"iot", "internet", "things", "smart", "device", "sensor", "automation",
		"biotech", "pharma", "pharmaceutical", "medical", "device", "diagnostic", "therapeutic",
		"green", "sustainable", "renewable", "solar", "wind", "energy", "environmental",
	}

	// Common English words (for compound domain segmentation)
	commonWords := []string{
		"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by", "from",
		"green", "blue", "red", "yellow", "black", "white", "gray", "brown", "orange", "purple",
		"grape", "grapes", "apple", "fruit", "berry", "berries",
		"as", "is", "are", "was", "were", "be", "been", "being", "have", "has", "had", "do", "does", "did",
		"will", "would", "should", "could", "may", "might", "must", "can",
		"this", "that", "these", "those", "it", "its", "they", "them", "their", "there", "then", "than",
		"all", "each", "every", "some", "any", "many", "much", "more", "most", "few", "little",
		"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten",
		"first", "second", "third", "last", "next", "previous", "new", "old", "young",
		"big", "small", "large", "tiny", "huge", "giant", "mini", "micro", "mega", "ultra",
		"good", "bad", "best", "worst", "better", "worse", "great", "excellent", "perfect",
		"fast", "slow", "quick", "rapid", "swift", "easy", "hard", "difficult", "simple", "complex",
		"high", "low", "top", "bottom", "up", "down", "left", "right", "north", "south", "east", "west",
		"red", "green", "blue", "yellow", "black", "white", "gray", "brown", "orange", "purple",
		"hot", "cold", "warm", "cool", "freeze", "melt", "fire", "water", "ice", "snow",
		"day", "night", "morning", "afternoon", "evening", "today", "tomorrow", "yesterday",
		"time", "hour", "minute", "second", "week", "month", "year", "date", "now", "then",
		"here", "where", "when", "why", "how", "what", "who", "which", "whose",
		"get", "got", "give", "gave", "take", "took", "make", "made", "do", "did", "go", "went", "come", "came",
		"see", "saw", "look", "watch", "read", "write", "speak", "talk", "say", "said", "tell", "told",
		"know", "knew", "think", "thought", "feel", "felt", "want", "need", "like", "love", "hate",
		"work", "job", "career", "business", "company", "office", "home", "house", "place", "space",
		"people", "person", "man", "woman", "child", "children", "family", "friend", "group", "team",
		"world", "country", "city", "town", "street", "road", "place", "location", "area", "region",
		"way", "path", "route", "direction", "method", "way", "means", "tool", "device", "machine",
		"thing", "item", "object", "stuff", "material", "product", "goods", "service", "help", "support",
		"problem", "issue", "question", "answer", "solution", "result", "outcome", "effect", "cause",
		"start", "begin", "end", "finish", "stop", "continue", "keep", "stay", "leave", "go",
		"open", "close", "turn", "move", "change", "become", "grow", "develop", "improve", "increase",
		"buy", "sell", "pay", "cost", "price", "money", "cash", "card", "bank", "account",
		"book", "read", "write", "page", "word", "letter", "number", "digit", "figure", "data",
		"computer", "phone", "mobile", "device", "screen", "keyboard", "mouse", "button", "click",
		"internet", "web", "site", "page", "link", "url", "email", "message", "text", "call",
		"music", "song", "sound", "voice", "listen", "hear", "play", "video", "movie", "film",
		"picture", "image", "photo", "camera", "take", "show", "display", "view", "see",
		"food", "eat", "drink", "meal", "breakfast", "lunch", "dinner", "snack", "taste", "cook",
		"car", "vehicle", "drive", "road", "street", "travel", "trip", "journey", "visit", "go",
		"school", "learn", "study", "teach", "student", "teacher", "class", "lesson", "test", "exam",
		"health", "sick", "doctor", "hospital", "medicine", "drug", "treatment", "care", "help",
		"sport", "game", "play", "team", "player", "win", "lose", "score", "goal", "point",
		"color", "shape", "size", "big", "small", "long", "short", "wide", "narrow", "tall",
		"happy", "sad", "angry", "excited", "tired", "sleepy", "awake", "alive", "dead", "live",
		"beautiful", "ugly", "pretty", "nice", "wonderful", "amazing", "awesome", "fantastic", "terrible",
	}

	// Add all words to dictionary
	for _, word := range businessTerms {
		dict[word] = true
	}
	for _, word := range industryTerms {
		dict[word] = true
	}
	for _, word := range commonWords {
		dict[word] = true
	}

	return dict
}

