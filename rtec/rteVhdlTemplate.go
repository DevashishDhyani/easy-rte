package rtec

import "text/template"

const rteVhdlTemplate = `{{define "_policyIn"}}{{$block := .}}
	--input policies
	{{range $polI, $pol := $block.Policies}}{{$pfbEnf := getPolicyEnfInfo $block $polI}}
	{{if not $pfbEnf}}//{{$pol.Name}} is broken!
	{{else}}{{/* this is where the policy comes in */}}--INPUT POLICY {{$pol.Name}} BEGIN 
		case {{$pol.Name}}_state is
			{{range $sti, $st := $pol.States}}when s_{{$pol.Name}}_{{$st.Name}}=>
				{{range $tri, $tr := $pfbEnf.InputPolicy.GetViolationTransitions}}{{if eq $tr.Source $st.Name}}{{/*
				*/}}
				if({{$cond := getVhdlECCTransitionCondition $block $tr.Condition}}{{$cond.IfCond}}) then
					--transition {{$tr.Source}} -> {{$tr.Destination}} on {{$tr.Condition}}
					--select a transition to solve the problem
					{{$solution := $pfbEnf.SolveViolationTransition $tr true}}
					{{if $solution.Comment}}--{{$solution.Comment}}{{end}}
					{{range $soleI, $sole := $solution.Expressions}}{{$sol := getVhdlECCTransitionCondition $block $sole}}{{$sol.IfCond}};
					{{end}}
				end if; {{end}}{{end}}
			{{end}}
		end if;
	{{end}}
	--INPUT POLICY {{$pol.Name}} END
	{{end}}
{{end}}

{{define "_policyOut"}}{{$block := .}}
	//output policies
	{{range $polI, $pol := $block.Policies}}{{$pfbEnf := getPolicyEnfInfo $block $polI}}
	{{if not $pfbEnf}}//{{$pol.Name}} is broken!
	{{else}}{{/* this is where the policy comes in */}}//OUTPUT POLICY {{$pol.Name}} BEGIN 
		switch(me->_policy_{{$pol.Name}}_state) {
			{{range $sti, $st := $pol.States}}case POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$st.Name}}:
				{{range $tri, $tr := $pfbEnf.OutputPolicy.GetViolationTransitions}}{{if eq $tr.Source $st.Name}}{{/*
				*/}}
				if({{$cond := getVhdlECCTransitionCondition $block $tr.Condition}}{{$cond.IfCond}}) {
					//transition {{$tr.Source}} -> {{$tr.Destination}} on {{$tr.Condition}}
					//select a transition to solve the problem
					{{$solution := $pfbEnf.SolveViolationTransition $tr false}}
					{{if $solution.Comment}}//{{$solution.Comment}}{{end}}
					{{range $soleI, $sole := $solution.Expressions}}{{$sol := getVhdlECCTransitionCondition $block $sole}}{{$sol.IfCond}};
					{{end}}
				} {{end}}{{end}}

				break;

			{{end}}
		}

		//advance timers
		{{range $varI, $var := $pfbEnf.OutputPolicy.GetDTimers}}
		me->{{$var.Name}}++;{{end}}

		//select transition to advance state
		switch(me->_policy_{{$pol.Name}}_state) {
			{{range $sti, $st := $pol.States}}case POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$st.Name}}:
				{{range $tri, $tr := $pfbEnf.OutputPolicy.GetNonViolationTransitions}}{{if eq $tr.Source $st.Name}}{{/*
				*/}}
				if({{$cond := getVhdlECCTransitionCondition $block $tr.Condition}}{{$cond.IfCond}}) {
					//transition {{$tr.Source}} -> {{$tr.Destination}} on {{$tr.Condition}}
					me->_policy_{{$pol.Name}}_state = POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$tr.Destination}};
					//set expressions
					{{range $exi, $ex := $tr.Expressions}}
					me->{{$ex.VarName}} = {{$ex.Value}};{{end}}
				} {{end}}{{end}}
				
				break;

			{{end}}
		}
	{{end}}
	//OUTPUT POLICY {{$pol.Name}} END
	{{end}}
{{end}}

{{define "functionVhdl"}}{{$block := index .Functions .FunctionIndex}}{{$blocks := .Functions}}
--This file should be called enforcer_{{$block.Name}}.vhdl
--This is autogenerated code. Edit by hand at your peril!

library ieee;
use ieee.std_logic_1164.all;
use ieee.numeric_std.all;

entity enforcer_{{$block.Name}} is
	port (
		{{range $index, $var := $block.InputVars}}
		{{$var.Name}}_ptc_in : in {{getVhdlType $var.Type}}{{if $var.InitialValue}} := "{{$var.InitialValue}}"{{end}};
		{{$var.Name}}_ptc_out : out {{getVhdlType $var.Type}}{{if $var.InitialValue}} := "{{$var.InitialValue}}"{{end}};
		{{end}}

		{{range $index, $var := $block.OutputVars}}
		{{$var.Name}}_ctp_in : in {{getVhdlType $var.Type}}{{if $var.InitialValue}} := "{{$var.InitialValue}}"{{end}};
		{{$var.Name}}_ctp_out : out {{getVhdlType $var.Type}}{{if $var.InitialValue}} := "{{$var.InitialValue}}"{{end}};
		{{end}}

		CLOCK: in std_logic
	);
end entity;

architecture rtl of enforcer_{{$block.Name}} is

--For each policy, we need a state_type for the state machine
{{range $polI, $pol := $block.Policies}}
	type {{$pol.Name}}_state_type is ({{if len $pol.States}}{{range $index, $state := $pol.States}}
			s_{{$pol.Name}}_{{$state}}{{if not $index}}, {{end}}{{end}}{{else}}s_{{$pol.Name}}_unknown{{end}}
	);

	signal {{$pol.Name}}_state : {{$pol.Name}}_state_type := {{if len $pol.States}}{{$state := index $pol.States 0}}s_{{$pol.Name}}_{{$state}}{{else}}s_{{$pol.Name}}_unknown{{end}};

	{{$pfbEnf := getPolicyEnfInfo $block $polI}}{{if not $pfbEnf}}--Policy is broken!{{else}}--internal vars
		{{range $vari, $var := $pfbEnf.OutputPolicy.InternalVars}}{{$var.Name}} : {{getVhdlType $var.Type}}{{if $var.InitialValue}} := "{{$var.InitialValue}}"{{end}};
	{{end}}{{end}}
{{end}}
begin

	process(CLOCK)
	{{range $index, $var := $block.InputVars}}
		{{$var.Name}} : {{getVhdlType $var.Type}}{{if $var.InitialValue}} := "{{$var.InitialValue}}"{{end}};
	{{end}}
	{{range $index, $var := $block.OutputVars}}
		{{$var.Name}} : {{getVhdlType $var.Type}}{{if $var.InitialValue}} := "{{$var.InitialValue}}"{{end}};
	{{end}}
	begin
		if (rising_edge(CLOCK)) then
			--capture synchronous inputs
		{{range $index, $var := $block.InputVars}}
			{{$var.Name}}_ptc := {{$var.Name}}_ptc_in;
		{{end}}

		{{if $block.Policies}}{{template "_policyIn" $block}}{{end}}
		
		end if;
	end process;
end architecture;{{end}}`

var vhdlTemplateFuncMap = template.FuncMap{
	"getVhdlECCTransitionCondition": getVhdlECCTransitionCondition,
	"getVhdlType":                   getVhdlType,
	"getPolicyEnfInfo":              getPolicyEnfInfo,
}

var vhdlTemplates = template.Must(template.New("").Funcs(vhdlTemplateFuncMap).Parse(rteVhdlTemplate))
