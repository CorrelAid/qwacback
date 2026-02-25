<?xml version="1.0" encoding="UTF-8"?>
<schema xmlns="http://purl.oclc.org/dsdl/schematron" queryBinding="xpath2">
    <ns prefix="ddi" uri="ddi:codebook:2_5"/>

    <pattern id="uniqueness">
        <rule context="ddi:var">
            <assert test="not(preceding::ddi:var/@ID = @ID)">Duplicate Variable ID found: <value-of select="@ID"/></assert>
        </rule>
        <rule context="var">
            <assert test="not(preceding::var/@ID = @ID)">Duplicate Variable ID found: <value-of select="@ID"/></assert>
        </rule>
        <rule context="ddi:varGrp">
            <assert test="not(preceding::ddi:varGrp/@ID = @ID)">Duplicate Variable Group ID found: <value-of select="@ID"/></assert>
        </rule>
        <rule context="varGrp">
            <assert test="not(preceding::varGrp/@ID = @ID)">Duplicate Variable Group ID found: <value-of select="@ID"/></assert>
        </rule>
    </pattern>

    <pattern id="essentials">
        <!-- Variable essentials -->
        <rule context="ddi:var">
            <assert test="@name">Variable <value-of select="@ID"/> is missing a name attribute.</assert>
            <assert test="@intrvl">Variable <value-of select="@name"/> is missing an intrvl attribute (use "discrete" or "contin").</assert>
            <assert test="ddi:qstn/@responseDomainType">Variable <value-of select="@name"/> is missing responseDomainType on qstn.</assert>
            <assert test="ddi:qstn/ddi:qstnLit">Variable <value-of select="@name"/> is missing a question literal (qstnLit).</assert>
            <assert test="ddi:varFormat">Variable <value-of select="@name"/> is missing technical format (varFormat).</assert>
            <assert test="ddi:concept and normalize-space(ddi:concept) != ''">Variable <value-of select="@name"/> is missing a concept element.</assert>
            <assert test="not(ddi:labl)">Variable <value-of select="@name"/> uses labl — use concept instead. labl is only for catgry elements.</assert>
        </rule>
        <rule context="var">
            <assert test="@name">Variable <value-of select="@ID"/> is missing a name attribute.</assert>
            <assert test="@intrvl">Variable <value-of select="@name"/> is missing an intrvl attribute (use "discrete" or "contin").</assert>
            <assert test="qstn/@responseDomainType">Variable <value-of select="@name"/> is missing responseDomainType on qstn.</assert>
            <assert test="qstn/qstnLit">Variable <value-of select="@name"/> is missing a question literal (qstnLit).</assert>
            <assert test="varFormat">Variable <value-of select="@name"/> is missing technical format (varFormat).</assert>
            <assert test="concept and normalize-space(concept) != ''">Variable <value-of select="@name"/> is missing a concept element.</assert>
            <assert test="not(labl)">Variable <value-of select="@name"/> uses labl — use concept instead. labl is only for catgry elements.</assert>
        </rule>

        <!-- Variable group essentials -->
        <rule context="ddi:varGrp">
            <assert test="@name">Variable Group <value-of select="@ID"/> is missing a name attribute.</assert>
            <assert test="@type = 'grid' or @type = 'multipleResp'">Variable Group <value-of select="@ID"/> has type="<value-of select="@type"/>". Only "grid" or "multipleResp" are supported.</assert>
            <assert test="ddi:concept and normalize-space(ddi:concept) != ''">Variable Group <value-of select="@ID"/> is missing a concept element.</assert>
            <assert test="not(ddi:labl)">Variable Group <value-of select="@ID"/> uses labl — use concept instead. labl is only for catgry elements.</assert>
        </rule>
        <rule context="varGrp">
            <assert test="@name">Variable Group <value-of select="@ID"/> is missing a name attribute.</assert>
            <assert test="@type = 'grid' or @type = 'multipleResp'">Variable Group <value-of select="@ID"/> has type="<value-of select="@type"/>". Only "grid" or "multipleResp" are supported.</assert>
            <assert test="concept and normalize-space(concept) != ''">Variable Group <value-of select="@ID"/> is missing a concept element.</assert>
            <assert test="not(labl)">Variable Group <value-of select="@ID"/> uses labl — use concept instead. labl is only for catgry elements.</assert>
        </rule>

        <!-- Category essentials -->
        <rule context="ddi:catgry">
            <assert test="ddi:catValu">A catgry element is missing catValu.</assert>
            <assert test="ddi:labl">A catgry element (value: <value-of select="ddi:catValu"/>) is missing a labl.</assert>
        </rule>
        <rule context="catgry">
            <assert test="catValu">A catgry element is missing catValu.</assert>
            <assert test="labl">A catgry element (value: <value-of select="catValu"/>) is missing a labl.</assert>
        </rule>
    </pattern>

    <pattern id="logic">
        <rule context="ddi:var">
            <assert test="not(ddi:qstn/@responseDomainType = 'category' or ddi:qstn/@responseDomainType = 'multiple') or ddi:catgry">
                Variable <value-of select="@name"/> (ID: <value-of select="@ID"/>) has a categorical response domain but no catgry elements.
            </assert>
        </rule>
        <rule context="var">
            <assert test="not(qstn/@responseDomainType = 'category' or qstn/@responseDomainType = 'multiple') or catgry">
                Variable <value-of select="@name"/> (ID: <value-of select="@ID"/>) has a categorical response domain but no catgry elements.
            </assert>
        </rule>

        <!-- Consistency Rule for Grids and Multiple Response groups -->
        <rule context="ddi:varGrp[@type='grid' or @type='multipleResp']">
            <assert test="every $id in tokenize(@var, '\s+') satisfies (not(//ddi:var[@ID=$id]/ddi:qstn/ddi:preQTxt) or normalize-space(//ddi:var[@ID=$id]/ddi:qstn/ddi:preQTxt) = normalize-space(ddi:txt))">
                Consistency Error: Variable Group <value-of select="@ID"/> (<value-of select="@type"/>) text does not match the preQTxt of its member variables.
            </assert>
            <!-- Multiple Choice specific: responseDomainType should be 'multiple' -->
            <assert test="not(@type='multipleResp') or (every $id in tokenize(@var, '\s+') satisfies (//ddi:var[@ID=$id]/ddi:qstn/@responseDomainType = 'multiple'))">
                Semantic Error: Variables in a multipleResp group (<value-of select="@ID"/>) should have responseDomainType="multiple".
            </assert>
            <!-- Grid specific: responseDomainType should be 'category' -->
            <assert test="not(@type='grid') or (every $id in tokenize(@var, '\s+') satisfies (//ddi:var[@ID=$id]/ddi:qstn/@responseDomainType = 'category'))">
                Semantic Error: Variables in a grid group (<value-of select="@ID"/>) should have responseDomainType="category".
            </assert>
        </rule>
        <rule context="varGrp[@type='grid' or @type='multipleResp']">
            <assert test="every $id in tokenize(@var, '\s+') satisfies (not(//var[@ID=$id]/qstn/preQTxt) or normalize-space(//var[@ID=$id]/qstn/preQTxt) = normalize-space(txt))">
                Consistency Error: Variable Group <value-of select="@ID"/> (<value-of select="@type"/>) text does not match the preQTxt of its member variables.
            </assert>
            <assert test="not(@type='multipleResp') or (every $id in tokenize(@var, '\s+') satisfies (//var[@ID=$id]/qstn/@responseDomainType = 'multiple'))">
                Semantic Error: Variables in a multipleResp group (<value-of select="@ID"/>) should have responseDomainType="multiple".
            </assert>
            <assert test="not(@type='grid') or (every $id in tokenize(@var, '\s+') satisfies (//var[@ID=$id]/qstn/@responseDomainType = 'category'))">
                Semantic Error: Variables in a grid group (<value-of select="@ID"/>) should have responseDomainType="category".
            </assert>
        </rule>
    </pattern>
</schema>
