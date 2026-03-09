#!/usr/bin/env python3
"""
Learning Desktop Research Content Setup
Creates 6 collections (directories) for skill tree content storage
"""

import json
from pathlib import Path

# Data directory for content storage
DATA_DIR = Path.home() / ".learning-desktop" / "research"
DATA_DIR.mkdir(parents=True, exist_ok=True)

# Output directory for collections
OUTPUT_DIR = DATA_DIR / "content"
OUTPUT_DIR.mkdir(exist_ok=True)

# Collection definitions
COLLECTIONS = [
    {
        "name": "character_manhood",
        "metadata": {
            "tree_id": "character-manhood",
            "category": "soft_skills",
            "target_audience": "young_men_13_17",
            "cultural_context": "north_miami_black_youth",
            "total_nodes": 9,
            "icon": "🦁",
            "description": "Character & Manhood: Build the foundation of who you are"
        },
        "nodes": [
            {"slug": "integrity", "title": "Integrity", "icon": "🤝", "position": {"row": 0, "col": 0}},
            {"slug": "responsibility", "title": "Ownership", "icon": "💪", "position": {"row": 1, "col": 0}},
            {"slug": "discipline", "title": "Self-Discipline", "icon": "⚔️", "position": {"row": 1, "col": 1}},
            {"slug": "respect", "title": "Respect", "icon": "👊", "position": {"row": 2, "col": 0}},
            {"slug": "resilience", "title": "Resilience", "icon": "🔥", "position": {"row": 2, "col": 1}},
            {"slug": "leadership", "title": "Leadership", "icon": "👑", "position": {"row": 3, "col": 0}},
            {"slug": "purpose", "title": "Purpose", "icon": "🎯", "position": {"row": 3, "col": 1}},
            {"slug": "brotherhood", "title": "Brotherhood", "icon": "🤝", "position": {"row": 4, "col": 0}},
            {"slug": "legacy", "title": "Legacy Thinking", "icon": "🏛️", "position": {"row": 4, "col": 1}},
        ]
    },
    {
        "name": "student_skills",
        "metadata": {
            "tree_id": "student-skills",
            "category": "soft_skills",
            "target_audience": "young_men_13_17",
            "cultural_context": "north_miami_black_youth",
            "total_nodes": 9,
            "icon": "📚",
            "description": "Student Skills: Master the art of learning"
        },
        "nodes": [
            {"slug": "focus", "title": "Deep Focus", "icon": "🎯", "position": {"row": 0, "col": 0}},
            {"slug": "note-taking", "title": "Note Taking", "icon": "📝", "position": {"row": 1, "col": 0}},
            {"slug": "time-management", "title": "Time Management", "icon": "⏰", "position": {"row": 1, "col": 1}},
            {"slug": "reading-strategy", "title": "Strategic Reading", "icon": "📖", "position": {"row": 2, "col": 0}},
            {"slug": "memory-techniques", "title": "Memory Systems", "icon": "🧠", "position": {"row": 2, "col": 1}},
            {"slug": "test-taking", "title": "Test Mastery", "icon": "✅", "position": {"row": 3, "col": 0}},
            {"slug": "research", "title": "Research Skills", "icon": "🔍", "position": {"row": 3, "col": 1}},
            {"slug": "teaching", "title": "Teaching Others", "icon": "👨‍🏫", "position": {"row": 4, "col": 0}},
            {"slug": "metacognition", "title": "Metacognition", "icon": "🔮", "position": {"row": 4, "col": 1}},
        ]
    },
    {
        "name": "entrepreneurship",
        "metadata": {
            "tree_id": "entrepreneurship",
            "category": "career",
            "target_audience": "young_men_13_17",
            "cultural_context": "north_miami_black_youth",
            "total_nodes": 9,
            "icon": "💼",
            "description": "Entrepreneurship: Build businesses, create value"
        },
        "nodes": [
            {"slug": "opportunity-spotting", "title": "See Opportunities", "icon": "👁️", "position": {"row": 0, "col": 0}},
            {"slug": "value-creation", "title": "Create Value", "icon": "🔨", "position": {"row": 1, "col": 0}},
            {"slug": "money-mindset", "title": "Money Mindset", "icon": "💰", "position": {"row": 1, "col": 1}},
            {"slug": "sales", "title": "Sales", "icon": "🤝", "position": {"row": 2, "col": 0}},
            {"slug": "marketing", "title": "Marketing", "icon": "📢", "position": {"row": 2, "col": 1}},
            {"slug": "financial-literacy", "title": "Financial Literacy", "icon": "📊", "position": {"row": 3, "col": 0}},
            {"slug": "legal-basics", "title": "Legal Foundation", "icon": "⚖️", "position": {"row": 3, "col": 1}},
            {"slug": "team-building", "title": "Build Teams", "icon": "👥", "position": {"row": 4, "col": 0}},
            {"slug": "scaling", "title": "Scale Up", "icon": "📈", "position": {"row": 4, "col": 1}},
        ]
    },
    {
        "name": "prompt_engineering",
        "metadata": {
            "tree_id": "prompt-engineering",
            "category": "technical",
            "target_audience": "young_men_13_17",
            "cultural_context": "north_miami_black_youth",
            "total_nodes": 6,
            "icon": "💬",
            "description": "Prompt Engineering: Master AI communication"
        },
        "nodes": [
            {"slug": "basic-prompts", "title": "Basic Prompt Writing", "icon": "📝", "position": {"row": 0, "col": 0}},
            {"slug": "chain-of-thought", "title": "Chain of Thought", "icon": "🔗", "position": {"row": 1, "col": 0}},
            {"slug": "few-shot-prompting", "title": "Few-Shot Prompting", "icon": "🎯", "position": {"row": 1, "col": 1}},
            {"slug": "advanced-techniques", "title": "Advanced Techniques", "icon": "🚀", "position": {"row": 2, "col": 0}},
            {"slug": "prompt-injection", "title": "Security & Safety", "icon": "🔒", "position": {"row": 2, "col": 1}},
            {"slug": "prompt-optimization", "title": "Prompt Optimization", "icon": "⚡", "position": {"row": 3, "col": 0}},
        ]
    },
    {
        "name": "ai_concepts",
        "metadata": {
            "tree_id": "ai-concepts",
            "category": "technical",
            "target_audience": "young_men_13_17",
            "cultural_context": "north_miami_black_youth",
            "total_nodes": 6,
            "icon": "🧠",
            "description": "AI Concepts: Understand how AI works"
        },
        "nodes": [
            {"slug": "llm-basics", "title": "How LLMs Work", "icon": "🧠", "position": {"row": 0, "col": 0}},
            {"slug": "embeddings", "title": "Embeddings & Vectors", "icon": "📊", "position": {"row": 1, "col": 0}},
            {"slug": "attention", "title": "Attention Mechanism", "icon": "👁️", "position": {"row": 1, "col": 1}},
            {"slug": "rag", "title": "RAG Systems", "icon": "📚", "position": {"row": 2, "col": 0}},
            {"slug": "fine-tuning", "title": "Fine-Tuning Models", "icon": "🎛️", "position": {"row": 2, "col": 1}},
            {"slug": "multimodal", "title": "Multimodal AI", "icon": "🖼️", "position": {"row": 3, "col": 0}},
        ]
    },
    {
        "name": "models_data",
        "metadata": {
            "tree_id": "models-data",
            "category": "technical",
            "target_audience": "young_men_13_17",
            "cultural_context": "north_miami_black_youth",
            "total_nodes": 6,
            "icon": "🗄️",
            "description": "Models & Data: Work with AI effectively"
        },
        "nodes": [
            {"slug": "model-selection", "title": "Model Selection", "icon": "🎯", "position": {"row": 0, "col": 0}},
            {"slug": "data-prep", "title": "Data Preparation", "icon": "📊", "position": {"row": 1, "col": 0}},
            {"slug": "vector-dbs", "title": "Vector Databases", "icon": "🗄️", "position": {"row": 1, "col": 1}},
            {"slug": "api-integration", "title": "API Integration", "icon": "🔌", "position": {"row": 2, "col": 0}},
            {"slug": "evaluation", "title": "Model Evaluation", "icon": "📈", "position": {"row": 2, "col": 1}},
            {"slug": "deployment", "title": "Deployment & Scaling", "icon": "🚀", "position": {"row": 3, "col": 0}},
        ]
    },
]

def setup_collections():
    """Create all content collections as directories"""
    created = []
    total_nodes = 0

    for collection_def in COLLECTIONS:
        name = collection_def["name"]
        metadata = collection_def["metadata"]
        nodes = collection_def.get("nodes", [])

        # Create collection directory
        collection_path = OUTPUT_DIR / name
        collection_path.mkdir(exist_ok=True)

        # Create metadata file
        metadata_file = collection_path / "_metadata.json"
        with open(metadata_file, 'w') as f:
            json.dump({
                "name": name,
                "metadata": metadata,
                "nodes": nodes,
                "created_at": str(collection_path.stat().st_ctime)
            }, f, indent=2)

        # Create node directories
        for node in nodes:
            node_path = collection_path / node["slug"]
            node_path.mkdir(exist_ok=True)
            # Create placeholder for topics
            (node_path / "_topics.json").write_text("[]")

        created.append(name)
        total_nodes += metadata["total_nodes"]
        print(f"✓ Created collection: {name} ({metadata['icon']} {metadata['tree_id']})")
        print(f"   Nodes: {len(nodes)}, Output: {collection_path}")

    print(f"\n🎉 Content setup complete!")
    print(f"   Location: {OUTPUT_DIR}")
    print(f"   Collections: {len(created)}")
    print(f"   Total nodes: {total_nodes}")

    return OUTPUT_DIR, created

if __name__ == "__main__":
    output_dir, collections = setup_collections()

    # Save collection info
    info_file = DATA_DIR / "collections.json"
    with open(info_file, 'w') as f:
        json.dump({
            "output_dir": str(output_dir),
            "collections": COLLECTIONS,
            "created_at": str(DATA_DIR.stat().st_ctime)
        }, f, indent=2)
    print(f"\n📋 Collection info saved to: {info_file}")
