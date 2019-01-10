package rtec

import (
	"text/template"

	"github.com/PRETgroup/goFB/goFB/stconverter"
)

const rteVerilogTemplate = `{{define "_policyIn"}}{{$block := .}}
	//input policies
	{{range $polI, $pol := $block.Policies}}{{$pfbEnf := getPolicyEnfInfo $block $polI}}
	{{if not $pfbEnf}}//{{$pol.Name}} is broken!
	{{else}}{{/* this is where the policy comes in */}}//INPUT POLICY {{$pol.Name}} BEGIN 
		case({{$block.Name}}_policy_{{$pol.Name}}_state)
		{{range $sti, $st := $pol.States}}` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$st.Name}}: begin
				{{range $tri, $tr := $pfbEnf.InputPolicy.GetViolationTransitions}}{{if eq $tr.Source $st.Name}}{{/*
				*/}}
				if ({{$cond := getVerilogECCTransitionCondition $block (compileExpression $tr.STGuard)}}{{$cond.IfCond}}) begin
					//transition {{$tr.Source}} -> {{$tr.Destination}} on {{$tr.Condition}}
					//select a transition to solve the problem
					{{$solution := $pfbEnf.SolveViolationTransition $tr true}}
					{{if $solution.Comment}}//{{$solution.Comment}}{{end}}
					{{range $soleI, $sole := $solution.Expressions}}{{$sol := getVerilogECCTransitionCondition $block (compileExpression $sole)}}{{$sol.IfCond}};
					{{end}}
				end{{end}}{{end}}
			end
			{{end}}
		endcase
	{{end}}
	//INPUT POLICY {{$pol.Name}} END
	{{end}}
{{end}}

{{define "_policyOut"}}{{$block := .}}
	//output policies
	{{range $polI, $pol := $block.Policies}}{{$pfbEnf := getPolicyEnfInfo $block $polI}}
	{{if not $pfbEnf}}//{{$pol.Name}} is broken!
	{{else}}{{/* this is where the policy comes in */}}//OUTPUT POLICY {{$pol.Name}} BEGIN 
		
		case({{$block.Name}}_policy_{{$pol.Name}}_state)
		{{range $sti, $st := $pol.States}}` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$st.Name}}: begin
				{{range $tri, $tr := $pfbEnf.OutputPolicy.GetViolationTransitions}}{{if eq $tr.Source $st.Name}}{{/*
				*/}}
				if ({{$cond := getVerilogECCTransitionCondition $block (compileExpression $tr.STGuard)}}{{$cond.IfCond}}) begin
					//transition {{$tr.Source}} -> {{$tr.Destination}} on {{$tr.Condition}}
					//select a transition to solve the problem
					{{$solution := $pfbEnf.SolveViolationTransition $tr false}}
					{{if $solution.Comment}}//{{$solution.Comment}}{{end}}
					{{range $soleI, $sole := $solution.Expressions}}{{$sol := getVerilogECCTransitionCondition $block (compileExpression $sole)}}{{$sol.IfCond}};
					{{end}}
				end {{end}}{{end}}
			end
			{{end}}
		endcase

		transTaken_{{$block.Name}}_policy_{{$pol.Name}} = 0;
		//select transition to advance state
		case({{$block.Name}}_policy_{{$pol.Name}}_state)
		{{range $sti, $st := $pol.States}}` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$st.Name}}: begin
				{{range $tri, $tr := $pfbEnf.OutputPolicy.GetTransitionsForSource $st.Name}}{{/*
				*/}}
				{{if $tri}}else {{end}}if ({{$cond := getVerilogECCTransitionCondition $block (compileExpression $tr.STGuard)}}{{$cond.IfCond}}) begin
					//transition {{$tr.Source}} -> {{$tr.Destination}} on {{$tr.Condition}}
					{{$block.Name}}_policy_{{$pol.Name}}_state = ` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$tr.Destination}};
					//set expressions
					{{range $exi, $ex := $tr.Expressions}}
					{{$ex.VarName}} = {{$ex.Value}};{{end}}
					transTaken_{{$block.Name}}_policy_{{$pol.Name}} = 1;
				end {{end}} else begin
					//only possible in a violation
					{{$block.Name}}_policy_{{$pol.Name}}_state = ` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_violation;
					transTaken_{{$block.Name}}_policy_{{$pol.Name}} = 1;
				end
			end
			{{end}}
		default begin
			//if we are here, we're in the violation state
			//the violation state permanently stays in violation
			{{$block.Name}}_policy_{{$pol.Name}}_state = ` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_violation;
			transTaken_{{$block.Name}}_policy_{{$pol.Name}} = 1;
		end
		endcase
	{{end}}
	//OUTPUT POLICY {{$pol.Name}} END
	{{end}}
{{end}}

{{define "functionVerilog"}}{{$block := index .Functions .FunctionIndex}}{{$blocks := .Functions}}
//This file should be called F_{{$block.Name}}.sv
//This is autogenerated code. Edit by hand at your peril!

//To check this file using EBMC, run the following command:
//$ ebmc F_{{$block.Name}}.sv

//For each policy, we need define types for the state machines
{{range $polI, $pol := $block.Policies}}
{{if len $pol.States}}{{range $index, $state := $pol.States}}
` + "`" + `define POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$state}} {{$index}}{{end}}{{else}}POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_unknown 0{{end}}
` + "`" + `define POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_violation {{if len $pol.States}}{{len $pol.States}}{{else}}1{{end}}
{{end}}

module F_combinatorialVerilog_{{$block.Name}} (
	//inputs (plant to controller){{range $index, $var := $block.InputVars}}
	input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_in,
	output wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_out,
	{{end}}
	//outputs (controller to plant){{range $index, $var := $block.OutputVars}}
	input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_in,
	output wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_out,
	{{end}}

	{{range $polI, $pol := $block.Policies}}//internal vars for EBMC to overload
	{{range $vari, $var := $pol.InternalVars}}{{if not $var.Constant}}input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_in,
	output wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_out,
	{{end}}{{end}}{{end}}
	
	//state input var
	{{range $polI, $pol := $block.Policies}}{{if $polI}},
	{{end}}input wire {{getVerilogWidthArray (add (len $pol.States) 1)}} {{$block.Name}}_policy_{{$pol.Name}}_state_in,
	output wire {{getVerilogWidthArray (add (len $pol.States) 1)}} {{$block.Name}}_policy_{{$pol.Name}}_state_out
	{{end}}
	
);

{{range $index, $var := $block.InputVars}}
{{getVerilogType $var.Type}} {{$var.Name}} {{if $var.InitialValue}}/* = {{$var.InitialValue}}*/{{end}};
{{end}}{{range $index, $var := $block.OutputVars}}
{{getVerilogType $var.Type}} {{$var.Name}} {{if $var.InitialValue}}/* = {{$var.InitialValue}}*/{{end}};
{{end}}{{range $polI, $pol := $block.Policies}}
{{$pfbEnf := getPolicyEnfInfo $block $polI}}{{if not $pfbEnf}}//Policy is broken!{{else}}//internal vars
{{range $vari, $var := $pfbEnf.OutputPolicy.InternalVars}}{{if $var.Constant}}localparam {{$var.Name}} = {{$var.InitialValue}}{{else}}{{getVerilogType $var.Type}} {{$var.Name}}{{if $var.InitialValue}}/* = {{$var.InitialValue}}*/{{end}}{{end}};
{{end}}{{end}}{{end}}

//For each policy, we need a reg for the state machine
{{range $polI, $pol := $block.Policies}}reg {{getVerilogWidthArray (add (len $pol.States) 1)}} {{$block.Name}}_policy_{{$pol.Name}}_state;
reg transTaken_{{$block.Name}}_policy_{{$pol.Name}}; //EBMC liveness check register flag (will be optimised away in normal compiles)
{{end}}

always @* begin

	{{range $index, $var := $block.InputVars}}
	{{$var.Name}} = {{$var.Name}}_ptc_in;
	{{end}}

	{{range $index, $var := $block.OutputVars}}
	{{$var.Name}} = {{$var.Name}}_ctp_in;
	{{end}}

	//capture state/time inputs
	{{range $polI, $pol := $block.Policies}}
	{{$block.Name}}_policy_{{$pol.Name}}_state =  {{$block.Name}}_policy_{{$pol.Name}}_state_in;
	{{$pfbEnf := getPolicyEnfInfo $block $polI}}{{if not $pfbEnf}}//Policy is broken!{{else}}//internal vars
	{{range $vari, $var := $pfbEnf.OutputPolicy.InternalVars}}{{if not $var.Constant}}{{$var.Name}} = {{$var.Name}}_in{{if $var.IsDTimer}} + 1{{end}};{{end}}
	{{end}}{{end}}{{end}}


	{{if $block.Policies}}{{template "_policyIn" $block}}{{end}}
	{{if $block.Policies}}{{template "_policyOut" $block}}{{end}}
end

{{range $index, $var := $block.InputVars}}
assign {{$var.Name}}_ptc_out = {{$var.Name}};
{{end}}

{{range $index, $var := $block.OutputVars}}
assign {{$var.Name}}_ctp_out = {{$var.Name}};
{{end}}

//emit state/time inputs
{{range $polI, $pol := $block.Policies}}
assign {{$block.Name}}_policy_{{$pol.Name}}_state_out =  {{$block.Name}}_policy_{{$pol.Name}}_state;
{{$pfbEnf := getPolicyEnfInfo $block $polI}}{{if not $pfbEnf}}//Policy is broken!{{else}}//internal vars
{{range $vari, $var := $pfbEnf.OutputPolicy.InternalVars}}{{if not $var.Constant}}assign {{$var.Name}}_out = {{$var.Name}};{{end}}
{{end}}{{end}}{{end}}

//For each policy, ensure correctness (systemverilog only) and liveness
{{range $polI, $pol := $block.Policies}}assert property ({{$block.Name}}_policy_{{$pol.Name}}_state_in >= ` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_violation || {{$block.Name}}_policy_{{$pol.Name}}_state_out != ` + "`" + `POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_violation);
//(temporarily disabled) assert property (transTaken_{{$block.Name}}_policy_{{$pol.Name}} == 1);
{{end}}

endmodule


module F_{{$block.Name}} (
	//inputs (plant to controller){{range $index, $var := $block.InputVars}}
	input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_in,
	output wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_out,
	{{end}}
	//outputs (controller to plant){{range $index, $var := $block.OutputVars}}
	input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_in,
	output wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_out,
	{{end}}

	//state var for EBMC to overload
	{{range $polI, $pol := $block.Policies}}input wire {{getVerilogWidthArray (add (len $pol.States) 1)}} {{$block.Name}}_policy_{{$pol.Name}}_state_embc_in,
	{{end}}
	{{range $polI, $pol := $block.Policies}}//internal vars for EBMC to overload
	{{range $vari, $var := $pol.InternalVars}}{{if not $var.Constant}}input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_embc_in,
	{{end}}{{end}}
	{{end}}
	

	input wire CLOCK
);

//For each policy, we need a reg for the state machine
{{range $polI, $pol := $block.Policies}}reg {{getVerilogWidthArray (add (len $pol.States) 1)}} {{$block.Name}}_policy_{{$pol.Name}}_state = 0;
wire {{getVerilogWidthArray (add (len $pol.States) 1)}} {{$block.Name}}_policy_{{$pol.Name}}_state_next;
{{end}}

/*{{range $index, $var := $block.InputVars}}
{{getVerilogType $var.Type}} {{$var.Name}} {{if $var.InitialValue}} = {{$var.InitialValue}}{{end}};
{{end}}{{range $index, $var := $block.OutputVars}}
{{getVerilogType $var.Type}} {{$var.Name}} {{if $var.InitialValue}} = {{$var.InitialValue}}{{end}};
{{end}}*/

{{range $polI, $pol := $block.Policies}}
{{$pfbEnf := getPolicyEnfInfo $block $polI}}{{if not $pfbEnf}}//Policy is broken!{{else}}//internal vars
{{range $vari, $var := $pfbEnf.OutputPolicy.InternalVars}}{{if not $var.Constant}}{{getVerilogType $var.Type}} {{$var.Name}}{{if $var.InitialValue}} = {{$var.InitialValue}}{{end}};
wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_next;
{{end}}{{end}}{{end}}{{end}}

F_combinatorialVerilog_{{$block.Name}} combPart (
	//inputs (plant to controller){{range $index, $var := $block.InputVars}}
	.{{$var.Name}}_ptc_in({{$var.Name}}_ptc_in),
	.{{$var.Name}}_ptc_out({{$var.Name}}_ptc_out),
	{{end}}
	//outputs (controller to plant){{range $index, $var := $block.OutputVars}}
	.{{$var.Name}}_ctp_in({{$var.Name}}_ctp_in),
	.{{$var.Name}}_ctp_out({{$var.Name}}_ctp_out),
	{{end}}

	{{range $polI, $pol := $block.Policies}}//internal vars for EBMC to overload
	{{range $vari, $var := $pol.InternalVars}}{{if not $var.Constant}}.{{$var.Name}}_in({{$var.Name}}),
	.{{$var.Name}}_out({{$var.Name}}_next),
	{{end}}{{end}}{{end}}
	
	//state input var
	{{range $polI, $pol := $block.Policies}}{{if $polI}},
	{{end}}.{{$block.Name}}_policy_{{$pol.Name}}_state_in({{$block.Name}}_policy_{{$pol.Name}}_state),
	.{{$block.Name}}_policy_{{$pol.Name}}_state_out({{$block.Name}}_policy_{{$pol.Name}}_state_next)
	{{end}}
);

	always@(posedge CLOCK) begin
		//capture synchronous inputs
	
		{{range $polI, $pol := $block.Policies}}//internal vars
		{{$block.Name}}_policy_{{$pol.Name}}_state = {{$block.Name}}_policy_{{$pol.Name}}_state_next;
		{{range $vari, $var := $pol.InternalVars}}{{if not $var.Constant}}{{$var.Name}} = {{$var.Name}}_next;
		{{end}}{{end}}{{end}}

		
	end
	
endmodule

module test_F_{{$block.Name}} (
	//inputs (plant to controller){{range $index, $var := $block.InputVars}}
	input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_in,
	output wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_out,
	{{end}}
	//outputs (controller to plant){{range $index, $var := $block.OutputVars}}
	input wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_in,
	output wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_out,
	{{end}}

	input wire CLOCK
);

{{range $index, $var := $block.InputVars}}
reg {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_in_reg;
wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_out_wire;
reg {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ptc_out_reg;
{{end}}

//outputs (controller to plant){{range $index, $var := $block.OutputVars}}
reg {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_in_reg;
wire {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_out_wire;
reg {{getVerilogWidthArrayForType $var.Type}} {{$var.Name}}_ctp_out_reg;
{{end}}

F_{{$block.Name}} test (

	//inputs (plant to controller){{range $index, $var := $block.InputVars}}
	.{{$var.Name}}_ptc_in({{$var.Name}}_ptc_in_reg),
	.{{$var.Name}}_ptc_out({{$var.Name}}_ptc_out_wire),
	{{end}}
	//outputs (controller to plant){{range $index, $var := $block.OutputVars}}
	.{{$var.Name}}_ctp_in({{$var.Name}}_ctp_in_reg),
	.{{$var.Name}}_ctp_out({{$var.Name}}_ctp_out_wire),
	{{end}}

	.CLOCK(CLOCK)
);

always@(posedge CLOCK) begin
	//inputs (plant to controller)
	{{range $index, $var := $block.InputVars}}
	{{$var.Name}}_ptc_in_reg = {{$var.Name}}_ptc_in;
	{{$var.Name}}_ptc_out_reg = {{$var.Name}}_ptc_out_wire;
	{{end}}

	//outputs (controller to plant){{range $index, $var := $block.OutputVars}}
	{{$var.Name}}_ctp_in_reg = {{$var.Name}}_ctp_in;
	{{$var.Name}}_ctp_out_reg = {{$var.Name}}_ctp_out_wire;
	{{end}}

end

//inputs (plant to controller)
{{range $index, $var := $block.InputVars}}
assign {{$var.Name}}_ptc_out = {{$var.Name}}_ptc_out_reg;
{{end}}

//outputs (controller to plant)
{{range $index, $var := $block.OutputVars}}
assign {{$var.Name}}_ctp_out = {{$var.Name}}_ctp_out_reg;
{{end}}

endmodule

{{end}}
`

var verilogTemplateFuncMap = template.FuncMap{
	"getVerilogECCTransitionCondition": getVerilogECCTransitionCondition,
	"getVerilogType":                   getVerilogType,
	"getPolicyEnfInfo":                 getPolicyEnfInfo,
	"getVerilogWidthArray":             getVerilogWidthArray,
	"getVerilogWidthArrayForType":      getVerilogWidthArrayForType,
	"add1IfClock":                      add1IfClock,

	"compileExpression": stconverter.VerilogCompileExpression,

	"add": add,
}

var verilogTemplates = template.Must(template.New("").Funcs(verilogTemplateFuncMap).Parse(rteVerilogTemplate))
