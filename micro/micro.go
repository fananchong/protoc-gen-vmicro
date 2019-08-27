package micro

import (
	"fmt"
	"strings"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath = "context"
	clientPkgPath  = "github.com/fananchong/v-micro/client"
	serverPkgPath  = "github.com/fananchong/v-micro/server"
)

func init() {
	generator.RegisterPlugin(new(micro))
}

// micro is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for v-micro support.
type micro struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "vmicro".
func (g *micro) Name() string {
	return "vmicro"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	clientPkg  string
	serverPkg  string
)

// Init initializes the plugin.
func (g *micro) Init(gen *generator.Generator) {
	g.gen = gen
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *micro) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *micro) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *micro) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *micro) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	contextPkg = string(g.gen.AddImport(contextPkgPath))
	clientPkg = string(g.gen.AddImport(clientPkgPath))
	serverPkg = string(g.gen.AddImport(serverPkgPath))

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *micro) GenerateImports(file *generator.FileDescriptor) {
}

func unexport(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// generateService generates all the code for the named service.
func (g *micro) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	path := fmt.Sprintf("6,%d", index) // 6 means service.

	origServName := service.GetName()
	serviceName := strings.ToLower(service.GetName())
	if pkg := file.GetPackage(); pkg != "" {
		serviceName = pkg
	}
	servName := generator.CamelCase(origServName)
	servAlias := servName + "Service"

	// strip suffix
	if strings.HasSuffix(servAlias, "ServiceService") {
		servAlias = strings.TrimSuffix(servAlias, "Service")
	}

	g.P()
	g.P("// Client API for ", servName, " service")
	g.P()

	// Client interface.
	g.P("type ", servAlias, " interface {")
	for i, method := range service.Method {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		g.P(g.generateClientSignature(servName, method))
	}
	g.P("}")
	g.P()
	g.P("type ", servName, "Callback interface {")
	for _, method := range service.Method {
		g.P(g.generateClientCallbackSignature(servName, method))
	}
	g.P("}")
	g.P()

	// Client structure.
	g.P("type ", unexport(servAlias), " struct {")
	g.P("c ", clientPkg, ".Client")
	g.P("name string")
	g.P("}")
	g.P()

	// NewClient factory.
	g.P("func New", servAlias, " (name string, hdcb ", servName, "Callback, c ", clientPkg, ".Client) ", servAlias, " {")
	g.P("if c == nil {")
	g.P("panic(\"client is nil\")")
	g.P("}")
	g.P("if len(name) == 0 {")
	g.P("panic(\"name is nil\")")
	g.P("}")
	g.P("if err := c.Handle(hdcb); err != nil {")
	g.P("panic(err)")
	g.P("}")
	g.P("return &", unexport(servAlias), "{")
	g.P("c: c,")
	g.P("name: name,")
	g.P("}")
	g.P("}")
	g.P()
	var methodIndex int
	serviceDescVar := "_" + servName + "_serviceDesc"
	// Client method implementations.
	for _, method := range service.Method {
		descExpr := fmt.Sprintf("&%s.Methods[%d]", serviceDescVar, methodIndex)
		methodIndex++
		g.generateClientMethod(serviceName, servName, serviceDescVar, method, descExpr)
	}

	g.P("// Server API for ", servName, " service")
	g.P()

	// Server interface.
	serverType := servName + "Handler"
	g.P("type ", serverType, " interface {")
	for i, method := range service.Method {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		g.P(g.generateServerSignature(servName, method))
	}
	g.P("}")
	g.P()

	// Server registration.
	g.P("func Register", servName, "Handler(s ", serverPkg, ".Server, hdlr ", serverType, ") error {")
	g.P("return s.Handle(hdlr)")
	g.P("}")
	g.P()
}

// generateClientSignature returns the client-side signature for a method.
func (g *micro) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	reqArg := ", req *" + g.typeName(method.GetInputType())
	return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) error", methName, contextPkg, reqArg, clientPkg)
}

func (g *micro) generateClientCallbackSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	var args []string
	ret := ""
	args = append(args, "ctx "+contextPkg+".Context")
	args = append(args, "req *"+g.typeName(method.GetInputType()))
	args = append(args, "rsp *"+g.typeName(method.GetOutputType()))
	return methName + "(" + strings.Join(args, ", ") + ") " + ret
}

func (g *micro) generateClientMethod(reqServ, servName, serviceDescVar string, method *pb.MethodDescriptorProto, descExpr string) {
	reqMethod := fmt.Sprintf("%s.%s", servName, method.GetName())
	servAlias := servName + "Service"
	// strip suffix
	if strings.HasSuffix(servAlias, "ServiceService") {
		servAlias = strings.TrimSuffix(servAlias, "Service")
	}
	g.P("func (c *", unexport(servAlias), ") ", g.generateClientSignature(servName, method), "{")
	g.P(`r := c.c.NewRequest(c.name, "`, reqMethod, `", req)`)
	g.P("err := ", `c.c.Call(ctx, r, opts...)`)
	g.P("if err != nil { return err }")
	g.P("return nil")
	g.P("}")
	g.P()
}

// generateServerSignature returns the server-side signature for a method.
func (g *micro) generateServerSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	var reqArgs []string
	ret := "error"
	reqArgs = append(reqArgs, "ctx "+contextPkg+".Context")
	reqArgs = append(reqArgs, "req *"+g.typeName(method.GetInputType()))
	reqArgs = append(reqArgs, "rsp *"+g.typeName(method.GetOutputType()))
	return methName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
}

// AddPluginToParams Simplify the protoc call statement by adding 'plugins=vmicro' directly to the command line arguments.
func AddPluginToParams(p string) string {
	params := p
	if strings.Contains(params, "plugins=") {
		params = strings.Replace(params, "plugins=", "plugins=vmicro+", -1)
	} else {
		if len(params) > 0 {
			params += ","
		}
		params += "plugins=vmicro"
	}
	return params
}
