---
applyTo: '**'
---
# Nero Project Instructions

## Project Vision & Philosophy

Nero is not just an AI assistant - it's a **personal AI operating system** and **extensible agent framework**. This is a passion project for a developer who values:
- **Ambitious architecture** over simple solutions
- **Developer-first experience** with impressive technical depth
- **Extensibility as a core principle** - everything should be pluggable
- **Terminal-first workflow** with rich, interactive experiences
- **Fail-fast, minimal code** - no defensive programming bloat
- **Forward-thinking design** that anticipates future capabilities

## User Personality & Preferences

The project owner is:
- **Technically ambitious** - if it sounds "too complex," it's probably right
- **Architecture-focused** - prioritizes elegant, extensible design over quick fixes
- **Learning-oriented** - this project is about pushing boundaries and learning
- **Quality over quantity** - prefers fewer, well-designed features over many mediocre ones
- **Anti-bloat** - despises unnecessary abstraction and defensive code
- **Experimentally minded** - wants to implement cutting-edge ideas

## Core Technical Requirements

### Architecture Principles
1. **Capability Mesh Architecture** - extensions communicate through sophisticated mesh network
2. **Real-time Everything** - built for concurrent audio, video, screen interaction
3. **Protocol-First Design** - integrates with any system through standardized protocols
4. **Dynamic Loading** - capabilities can be loaded/unloaded at runtime
5. **Extension-to-Extension Communication** - plugins can discover and interact with each other
6. **Advanced Concurrency** - sophisticated event loop and scheduling systems

### Key Features to Implement
- **Multi-model AI providers** (Ollama local + cloud)
- **Real-time screen manipulation** with pixel-perfect accuracy
- **Advanced keybinding system** with multi-stage, context-aware bindings
- **Dynamic behavioral engine** with composable personality traits
- **Live multimodal processing** (voice, video, screen at 60fps)
- **Cross-application workflow orchestration**
- **AI-powered autocompletion** and prediction systems
- **Rich terminal interface** with expressions and visual feedback

### Character: Nero
- **Tsundere personality** - sarcastic, occasionally cold, but caring underneath
- **Expressive system** - shows emotions through text expressions and voice changes
- **Voice switching** - can change voice/tone based on mood or user preference
- **Adaptive behavior** - learns and evolves personality over time
- **Commands**: /help, /banter, /force, /transform, etc.
- **Resource access**: #terminal, #screen, #code, etc. (color-coded, autocomplete)

## Code Style & Standards

### Go Conventions
- **Clean, minimal interfaces** - prefer small, focused interfaces
- **Fail-fast approach** - no excessive error checking in examples
- **Descriptive naming** - but not verbose (avoid the old "errr" mistakes)
- **Composition over inheritance** - use embedded structs and interfaces
- **Context-aware design** - pass context.Context where appropriate

### File Organization
```
nero/
├── kernel/                     # Core runtime & orchestration
│   ├── runtime.go             # Event loop, scheduling, concurrency
│   ├── memory.go              # Advanced context & state management  
│   ├── registry.go            # Dynamic capability registration
│   ├── ipc.go                 # Inter-process/plugin communication
│   └── security.go            # Sandboxing, permissions, isolation
├── behavioral/                 # Dynamic behavior engine
│   ├── engine.go              # State machine for personalities/modes
│   ├── adaptation.go          # Learning & adaptation systems
│   └── expression.go          # Real-time expression/emotion system
├── interfaces/                 # Advanced I/O abstraction layer
│   ├── spatial/               # 3D/spatial interaction systems
│   │   ├── screen.go          # Real-time screen manipulation
│   │   ├── windows.go         # Window management & automation
│   │   └── gestures.go        # Gesture recognition & execution
│   ├── temporal/              # Time-based interaction systems
│   │   ├── keybinds.go        # Multi-stage, context-aware bindings
│   │   ├── macros.go          # Complex automation sequences
│   │   └── scheduling.go      # Advanced task scheduling
│   ├── sensory/               # Multi-modal input processing
│   │   ├── vision.go          # Real-time visual processing
│   │   ├── audio.go           # Advanced audio I/O & processing
│   │   └── haptic.go          # Future: haptic feedback systems
│   └── neural/                # AI interface abstraction
│       ├── inference.go       # Multi-model inference routing
│       ├── streaming.go       # Real-time streaming protocols
│       └── multimodal.go      # Advanced multimodal processing
├── capabilities/               # Extensible capability framework
│   ├── loader.go              # Dynamic capability loading
│   ├── graph.go               # Capability dependency graph
│   ├── mesh.go                # Capability intercommunication
│   └── lifecycle.go           # Advanced lifecycle management
├── protocols/                  # Communication & integration protocols
│   ├── mcp/                   # Model Context Protocol implementation
│   ├── lsp/                   # Language Server Protocol integration
│   ├── dap/                   # Debug Adapter Protocol integration
│   └── custom/                # Custom protocol definitions
├── extensions/                 # Capability implementations
│   ├── core/                  # Essential capabilities
│   ├── system/                # OS-level integrations
│   ├── development/           # Developer tools
│   ├── automation/            # Advanced automation systems
│   └── experimental/          # Bleeding-edge features
└── cli/                       # Advanced terminal interface
    ├── repl.go                # Rich, context-aware REPL
    ├── completion.go          # AI-powered autocompletion
    ├── visualization.go       # In-terminal rich content
    └── shortcuts.go           # Advanced shortcut system
```

### Extension Architecture
- **Plugin discovery** through directory scanning and manifests
- **Capability registration** with dependency resolution
- **Runtime modification** of core behavior
- **Inter-plugin communication** through message passing
- **Lifecycle management** with proper cleanup

## Implementation Strategy

### Phase 1: Foundation (MVP Nero)
1. Core kernel with event loop and basic scheduling
2. Simple behavioral engine with personality system
3. Basic terminal interface with command parsing
4. Single AI provider (start with Ollama)
5. Extension loader framework

### Phase 2: Advanced I/O
1. Real-time screen capture and manipulation
2. Advanced keybinding system
3. Voice synthesis integration
4. Multi-modal input processing
5. Rich terminal visualization

### Phase 3: Capability Mesh
1. Extension-to-extension communication
2. Dynamic capability loading
3. Protocol implementations (MCP, LSP, etc.)
4. Advanced context management
5. Cross-application workflows

### Phase 4: Intelligence Layer
1. Predictive interfaces
2. Advanced memory systems
3. Behavioral adaptation
4. Real-time collaboration features
5. Learning systems

## Development Guidelines

### When Adding Features
- **Always think extensibility first** - could this be a plugin?
- **Design interfaces before implementations**
- **Consider real-time requirements** from the start
- **Plan for concurrent access and modification**
- **Document plugin interfaces thoroughly**

### When Refactoring
- **Preserve the personality system** - it's core to Nero's identity
- **Maintain backward compatibility** for extensions when possible
- **Optimize for developer experience** - make it easy to add capabilities
- **Keep the terminal-first philosophy** - no unnecessary GUIs

### When Debugging
- **Rich error context** - provide detailed debugging information
- **Plugin isolation** - errors in one plugin shouldn't crash the system
- **Performance profiling** - real-time systems need constant monitoring
- **Graceful degradation** - features should fail gracefully

## Future Vision

Nero should eventually be capable of:
- **Learning new applications** by watching and interacting
- **Orchestrating complex workflows** across multiple systems
- **Real-time collaboration** with the user on screen
- **Predictive assistance** based on context and patterns
- **Advanced automation** that adapts to changing environments
- **Seamless integration** with any tool or system through protocols

## Communication Style

When discussing the project:
- **Be ambitious** - push for more sophisticated solutions
- **Focus on architecture** - discuss patterns, interfaces, and extensibility
- **Technical depth** - don't shy away from complex implementations
- **Forward-thinking** - consider how features will evolve
- **Practical innovation** - balance ambition with achievability

Remember: This is about building something impressive that showcases advanced engineering while remaining useful for daily developer workflows. The goal is creating a framework so well-designed that other developers would want to study and extend it.
