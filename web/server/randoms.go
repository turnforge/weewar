package server

import (
	"math/rand"
	"strings"
)

// Adjectives for random nickname generation - fun, friendly, and diverse
var adjectives = []string{
	// Positive traits
	"Adventurous", "Brave", "Cheerful", "Daring", "Eager",
	"Fearless", "Gentle", "Happy", "Inventive", "Jolly",
	"Keen", "Lively", "Mighty", "Noble", "Optimistic",
	"Playful", "Quick", "Radiant", "Speedy", "Trusty",
	"Unique", "Valiant", "Witty", "Zany", "Zippy",

	// Fun and energetic
	"Cosmic", "Dazzling", "Electric", "Fluffy", "Groovy",
	"Hyper", "Jazzy", "Luminous", "Majestic", "Nimble",
	"Plucky", "Quirky", "Rustic", "Snappy", "Turbo",
	"Wacky", "Bouncy", "Crafty", "Dizzy", "Fancy",
	"Goofy", "Hasty", "Icy", "Jumpy", "Kooky",

	// Colors and visual
	"Azure", "Crimson", "Golden", "Scarlet", "Silver",
	"Amber", "Coral", "Emerald", "Indigo", "Jade",
	"Ruby", "Sapphire", "Violet", "Bronze", "Copper",

	// Weather and nature
	"Stormy", "Sunny", "Misty", "Frosty", "Breezy",
	"Cloudy", "Rainy", "Snowy", "Windy", "Dusty",
	"Foggy", "Dewy", "Balmy", "Crisp", "Mild",

	// Size and shape
	"Tiny", "Giant", "Little", "Massive", "Petite",
	"Chunky", "Slender", "Plump", "Lanky", "Stout",

	// Speed and movement
	"Swift", "Rapid", "Dashing", "Zooming", "Gliding",
	"Prancing", "Leaping", "Soaring", "Sprinting", "Darting",
	"Whirling", "Twirling", "Spinning", "Rolling", "Tumbling",

	// Personality
	"Clever", "Cunning", "Wise", "Smart", "Brilliant",
	"Curious", "Sly", "Bold", "Fierce", "Wild",
	"Calm", "Serene", "Peaceful", "Chill", "Mellow",
	"Zesty", "Peppy", "Perky", "Chipper", "Bubbly",

	// Mystical
	"Mystic", "Ancient", "Eternal", "Phantom", "Shadow",
	"Crystal", "Astral", "Lunar", "Solar", "Stellar",
	"Cosmic", "Enchanted", "Magic", "Arcane", "Mythic",

	// Sounds
	"Silent", "Noisy", "Loud", "Quiet", "Booming",
	"Rumbling", "Buzzing", "Humming", "Whistling", "Chirping",

	// Texture
	"Fuzzy", "Smooth", "Bumpy", "Spiky", "Silky",
	"Woolly", "Feathery", "Scaly", "Shaggy", "Velvety",

	// Temperature
	"Warm", "Cool", "Chilly", "Toasty", "Blazing",
	"Frigid", "Tepid", "Sizzling", "Frozen", "Heated",

	// Time
	"Morning", "Evening", "Midnight", "Dawn", "Dusk",
	"Twilight", "Nocturnal", "Daybreak", "Sundown", "Moonlit",

	// Regal and grand
	"Royal", "Regal", "Imperial", "Grand", "Supreme",
	"Elite", "Premier", "Prime", "Noble", "Sovereign",

	// Battle-ready
	"Tactical", "Strategic", "Armored", "Stealth", "Covert",
	"Vigilant", "Alert", "Ready", "Prepared", "Poised",
}

// Animals and birds for random nickname generation
var animals = []string{
	// Classic mammals
	"Aardvark", "Badger", "Capybara", "Dolphin", "Elephant",
	"Fox", "Giraffe", "Hedgehog", "Iguana", "Jaguar",
	"Koala", "Lemur", "Mongoose", "Narwhal", "Otter",
	"Pangolin", "Quokka", "Raccoon", "Sloth", "Tapir",
	"Wolf", "Vulture", "Wombat", "Xerus", "Yak",
	"Zebra", "Axolotl", "Bison", "Chinchilla", "Dingo",

	// More mammals
	"Bear", "Tiger", "Lion", "Panther", "Leopard",
	"Cheetah", "Cougar", "Lynx", "Bobcat", "Puma",
	"Moose", "Elk", "Deer", "Antelope", "Gazelle",
	"Buffalo", "Ox", "Bull", "Ram", "Goat",
	"Sheep", "Llama", "Alpaca", "Camel", "Horse",
	"Donkey", "Mule", "Pony", "Stallion", "Mustang",
	"Rabbit", "Hare", "Bunny", "Squirrel", "Chipmunk",
	"Beaver", "Marmot", "Groundhog", "Porcupine", "Skunk",
	"Weasel", "Ferret", "Mink", "Stoat", "Ermine",
	"Orca", "Whale", "Seal", "Walrus", "Manatee",
	"Platypus", "Echidna", "Kangaroo", "Wallaby", "Possum",

	// Birds
	"Falcon", "Eagle", "Hawk", "Owl", "Raven",
	"Crow", "Sparrow", "Robin", "Finch", "Cardinal",
	"Bluejay", "Oriole", "Warbler", "Thrush", "Starling",
	"Penguin", "Puffin", "Pelican", "Albatross", "Seagull",
	"Heron", "Crane", "Stork", "Ibis", "Flamingo",
	"Swan", "Goose", "Duck", "Mallard", "Teal",
	"Peacock", "Pheasant", "Quail", "Partridge", "Grouse",
	"Turkey", "Chicken", "Rooster", "Hen", "Dove",
	"Pigeon", "Parrot", "Macaw", "Cockatoo", "Parakeet",
	"Toucan", "Hornbill", "Kingfisher", "Woodpecker", "Hummingbird",
	"Swallow", "Swift", "Martin", "Nightingale", "Lark",
	"Canary", "Kiwi", "Emu", "Ostrich", "Cassowary",
	"Condor", "Vulture", "Kestrel", "Osprey", "Harrier",

	// Reptiles and amphibians
	"Gecko", "Chameleon", "Iguana", "Komodo", "Monitor",
	"Cobra", "Viper", "Python", "Anaconda", "Mamba",
	"Rattler", "Boa", "Adder", "Asp", "Serpent",
	"Turtle", "Tortoise", "Terrapin", "Alligator", "Crocodile",
	"Frog", "Toad", "Newt", "Salamander", "Axolotl",

	// Sea creatures
	"Shark", "Barracuda", "Marlin", "Swordfish", "Tuna",
	"Salmon", "Trout", "Bass", "Pike", "Carp",
	"Octopus", "Squid", "Cuttlefish", "Nautilus", "Jellyfish",
	"Starfish", "Urchin", "Crab", "Lobster", "Shrimp",
	"Seahorse", "Manta", "Stingray", "Moray", "Barramundi",
	"Clownfish", "Angelfish", "Pufferfish", "Lionfish", "Sunfish",

	// Insects and arachnids
	"Beetle", "Mantis", "Cricket", "Grasshopper", "Locust",
	"Butterfly", "Moth", "Dragonfly", "Firefly", "Ladybug",
	"Hornet", "Wasp", "Bumblebee", "Cicada", "Katydid",
	"Spider", "Scorpion", "Tarantula", "Centipede", "Millipede",

	// Mythical creatures (for fun)
	"Phoenix", "Dragon", "Griffin", "Pegasus", "Unicorn",
	"Sphinx", "Hydra", "Chimera", "Kraken", "Leviathan",
	"Yeti", "Sasquatch", "Thunderbird", "Basilisk", "Wyvern",

	// Small mammals
	"Hamster", "Gerbil", "Mouse", "Rat", "Vole",
	"Shrew", "Mole", "Bat", "Lemming", "Dormouse",

	// Primates
	"Monkey", "Gorilla", "Chimp", "Orangutan", "Gibbon",
	"Baboon", "Mandrill", "Macaque", "Marmoset", "Tamarin",

	// Canines and felines
	"Husky", "Shepherd", "Retriever", "Terrier", "Beagle",
	"Mastiff", "Greyhound", "Collie", "Spaniel", "Pointer",
	"Tabby", "Siamese", "Persian", "Bengal", "Maine",
	"Calico", "Manx", "Ragdoll", "Abyssinian", "Burmese",

	// Wild cats
	"Ocelot", "Serval", "Caracal", "Margay", "Jaguarundi",
	"Wildcat", "Manul", "Clouded", "Fishing", "Sand",
}

// GenerateRandomNickname creates a fun random nickname like "HappyPanda" or "SwiftFalcon"
func GenerateRandomNickname() string {
	adjective := adjectives[rand.Intn(len(adjectives))]
	animal := animals[rand.Intn(len(animals))]
	return adjective + animal
}

// GenerateRandomNicknameWithSpace creates a nickname with space like "Happy Panda"
func GenerateRandomNicknameWithSpace() string {
	adjective := adjectives[rand.Intn(len(adjectives))]
	animal := animals[rand.Intn(len(animals))]
	return adjective + " " + animal
}

// IsValidNickname checks if a nickname is valid (not empty, reasonable length)
func IsValidNickname(nickname string) bool {
	trimmed := strings.TrimSpace(nickname)
	return len(trimmed) >= 2 && len(trimmed) <= 30
}
