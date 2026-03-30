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
            <assert test="count(ddi:notes) &lt;= 1">Variable <value-of select="@name"/> has multiple notes elements. Only one notes element per variable is allowed.</assert>
        </rule>
        <rule context="var">
            <assert test="@name">Variable <value-of select="@ID"/> is missing a name attribute.</assert>
            <assert test="@intrvl">Variable <value-of select="@name"/> is missing an intrvl attribute (use "discrete" or "contin").</assert>
            <assert test="qstn/@responseDomainType">Variable <value-of select="@name"/> is missing responseDomainType on qstn.</assert>
            <assert test="qstn/qstnLit">Variable <value-of select="@name"/> is missing a question literal (qstnLit).</assert>
            <assert test="varFormat">Variable <value-of select="@name"/> is missing technical format (varFormat).</assert>
            <assert test="concept and normalize-space(concept) != ''">Variable <value-of select="@name"/> is missing a concept element.</assert>
            <assert test="not(labl)">Variable <value-of select="@name"/> uses labl — use concept instead. labl is only for catgry elements.</assert>
            <assert test="count(notes) &lt;= 1">Variable <value-of select="@name"/> has multiple notes elements. Only one notes element per variable is allowed.</assert>
        </rule>

        <!-- Variable group essentials -->
        <rule context="ddi:varGrp">
            <assert test="@name">Variable Group <value-of select="@ID"/> is missing a name attribute.</assert>
            <assert test="@type = 'grid' or @type = 'multipleResp' or @type = 'other'">Variable Group <value-of select="@ID"/> has type="<value-of select="@type"/>". Only "grid", "multipleResp", or "other" are supported.</assert>
            <assert test="ddi:concept and normalize-space(ddi:concept) != ''">Variable Group <value-of select="@ID"/> is missing a concept element.</assert>
            <assert test="not(ddi:labl)">Variable Group <value-of select="@ID"/> uses labl — use concept instead. labl is only for catgry elements.</assert>
        </rule>
        <rule context="varGrp">
            <assert test="@name">Variable Group <value-of select="@ID"/> is missing a name attribute.</assert>
            <assert test="@type = 'grid' or @type = 'multipleResp' or @type = 'other'">Variable Group <value-of select="@ID"/> has type="<value-of select="@type"/>". Only "grid", "multipleResp", or "other" are supported.</assert>
            <assert test="concept and normalize-space(concept) != ''">Variable Group <value-of select="@ID"/> is missing a concept element.</assert>
            <assert test="not(labl)">Variable Group <value-of select="@ID"/> uses labl — use concept instead. labl is only for catgry elements.</assert>
        </rule>

        <!-- Category essentials -->
        <rule context="ddi:catgry">
            <assert test="ddi:catValu">A catgry element is missing catValu.</assert>
            <!-- labl is required for category (single-choice/grid) but not for multiple (checkboxes) -->
            <assert test="ddi:labl or ancestor::ddi:var/ddi:qstn/@responseDomainType = 'multiple'">A catgry element (value: <value-of select="ddi:catValu"/>) is missing a labl (required for responseDomainType="category").</assert>
        </rule>
        <rule context="catgry">
            <assert test="catValu">A catgry element is missing catValu.</assert>
            <assert test="labl or ancestor::var/qstn/@responseDomainType = 'multiple'">A catgry element (value: <value-of select="catValu"/>) is missing a labl (required for responseDomainType="category").</assert>
        </rule>
    </pattern>

    <pattern id="logic">
        <!--
            Categorical variables must have catgry elements UNLESS concept/@vocab
            is present, which signals that the categories are defined by an external
            standard code list (e.g. ISO 3166-1 country codes) and need not be
            repeated inline.
        -->
        <rule context="ddi:var">
            <assert test="not(ddi:qstn/@responseDomainType = 'category' or ddi:qstn/@responseDomainType = 'multiple') or ddi:catgry or ddi:concept/@vocab">
                Variable <value-of select="@name"/> (ID: <value-of select="@ID"/>) has a categorical response domain but no catgry elements (add categories or reference an external vocabulary via concept/@vocab).
            </assert>
        </rule>
        <rule context="var">
            <assert test="not(qstn/@responseDomainType = 'category' or qstn/@responseDomainType = 'multiple') or catgry or concept/@vocab">
                Variable <value-of select="@name"/> (ID: <value-of select="@ID"/>) has a categorical response domain but no catgry elements (add categories or reference an external vocabulary via concept/@vocab).
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
        <!-- Parent "other" varGrp rules (semi-open question hierarchy) -->
        <rule context="ddi:varGrp[@type='other']">
            <!-- Must reference at least one var or child varGrp -->
            <assert test="@var or @varGrp">
                Parent varGrp <value-of select="@ID"/> (type="other") must reference variables (@var) or child groups (@varGrp).
            </assert>
            <!-- Each referenced child varGrp must exist -->
            <assert test="not(@varGrp) or (every $id in tokenize(@varGrp, '\s+') satisfies //ddi:varGrp[@ID=$id])">
                Parent varGrp <value-of select="@ID"/> references a child varGrp that does not exist.
            </assert>
            <!-- Each referenced var must exist -->
            <assert test="not(@var) or (every $id in tokenize(@var, '\s+') satisfies //ddi:var[@ID=$id])">
                Parent varGrp <value-of select="@ID"/> references a variable that does not exist.
            </assert>
        </rule>
        <rule context="varGrp[@type='other']">
            <assert test="@var or @varGrp">
                Parent varGrp <value-of select="@ID"/> (type="other") must reference variables (@var) or child groups (@varGrp).
            </assert>
            <assert test="not(@varGrp) or (every $id in tokenize(@varGrp, '\s+') satisfies //varGrp[@ID=$id])">
                Parent varGrp <value-of select="@ID"/> references a child varGrp that does not exist.
            </assert>
            <assert test="not(@var) or (every $id in tokenize(@var, '\s+') satisfies //var[@ID=$id])">
                Parent varGrp <value-of select="@ID"/> references a variable that does not exist.
            </assert>
        </rule>
    </pattern>

    <pattern id="other_variables">
        <!--
            Rules for "_other" (semi-open / halb-offen) variables.
            Convention: a var whose @name ends in "_other" is a free-text
            specification field linked to a base variable or group by naming
            convention (e.g. geschlecht_other → geschlecht).
        -->

        <!-- Namespaced variant -->
        <rule context="ddi:var[ends-with(@name, '_other')]">
            <!-- _other vars must be text variables -->
            <assert test="ddi:qstn/@responseDomainType = 'text'">
                Variable <value-of select="@name"/> ends in "_other" but has responseDomainType="<value-of select="ddi:qstn/@responseDomainType"/>". Expected "text".
            </assert>
            <assert test="@intrvl = 'discrete'">
                Variable <value-of select="@name"/> ends in "_other" but has intrvl="<value-of select="@intrvl"/>". Expected "discrete".
            </assert>
            <assert test="ddi:varFormat/@type = 'character'">
                Variable <value-of select="@name"/> ends in "_other" but has varFormat type="<value-of select="ddi:varFormat/@type"/>". Expected "character".
            </assert>
            <!-- A matching base var or varGrp must exist -->
            <assert test="//ddi:var[@name = substring-before(current()/@name, '_other')] or //ddi:varGrp[@name = substring-before(current()/@name, '_other')]">
                Variable <value-of select="@name"/> ends in "_other" but no matching base variable or group named "<value-of select="substring-before(@name, '_other')"/>" was found.
            </assert>
            <!-- When base is a single-choice var, it must have catValu="other" (convention for round-trip) -->
            <assert test="not(//ddi:var[@name = substring-before(current()/@name, '_other')]) or //ddi:var[@name = substring-before(current()/@name, '_other')]/ddi:catgry[ddi:catValu = 'other'] or //ddi:varGrp[@name = substring-before(current()/@name, '_other')]">
                Variable <value-of select="@name"/>: the base variable "<value-of select="substring-before(@name, '_other')"/>" must have a catgry with catValu="other" (convention for round-trip conversion).
            </assert>
            <!-- _other var must NOT be listed in a multipleResp group's var attribute -->
            <assert test="not(//ddi:varGrp[@type='multipleResp' and contains(concat(' ', @var, ' '), concat(' ', current()/@ID, ' '))])">
                Variable <value-of select="@name"/> (text _other) must not be a member of a multipleResp group. It should be a standalone variable outside the group.
            </assert>
            <!-- _other and long list (concept/@vocab) are mutually exclusive -->
            <assert test="not(//ddi:var[@name = substring-before(current()/@name, '_other')]/ddi:concept/@vocab)">
                Variable <value-of select="@name"/>: the base variable "<value-of select="substring-before(@name, '_other')"/>" uses concept/@vocab (long list). Long list and _other cannot be combined.
            </assert>
        </rule>

        <!-- Non-namespaced variant -->
        <rule context="var[ends-with(@name, '_other')]">
            <assert test="qstn/@responseDomainType = 'text'">
                Variable <value-of select="@name"/> ends in "_other" but has responseDomainType="<value-of select="qstn/@responseDomainType"/>". Expected "text".
            </assert>
            <assert test="@intrvl = 'discrete'">
                Variable <value-of select="@name"/> ends in "_other" but has intrvl="<value-of select="@intrvl"/>". Expected "discrete".
            </assert>
            <assert test="varFormat/@type = 'character'">
                Variable <value-of select="@name"/> ends in "_other" but has varFormat type="<value-of select="varFormat/@type"/>". Expected "character".
            </assert>
            <assert test="//var[@name = substring-before(current()/@name, '_other')] or //varGrp[@name = substring-before(current()/@name, '_other')]">
                Variable <value-of select="@name"/> ends in "_other" but no matching base variable or group named "<value-of select="substring-before(@name, '_other')"/>" was found.
            </assert>
            <!-- When base is a single-choice var, it must have catValu="other" (convention for round-trip) -->
            <assert test="not(//var[@name = substring-before(current()/@name, '_other')]) or //var[@name = substring-before(current()/@name, '_other')]/catgry[catValu = 'other'] or //varGrp[@name = substring-before(current()/@name, '_other')]">
                Variable <value-of select="@name"/>: the base variable "<value-of select="substring-before(@name, '_other')"/>" must have a catgry with catValu="other" (convention for round-trip conversion).
            </assert>
            <assert test="not(//varGrp[@type='multipleResp' and contains(concat(' ', @var, ' '), concat(' ', current()/@ID, ' '))])">
                Variable <value-of select="@name"/> (text _other) must not be a member of a multipleResp group. It should be a standalone variable outside the group.
            </assert>
            <!-- _other and long list (concept/@vocab) are mutually exclusive -->
            <assert test="not(//var[@name = substring-before(current()/@name, '_other')]/concept/@vocab)">
                Variable <value-of select="@name"/>: the base variable "<value-of select="substring-before(@name, '_other')"/>" uses concept/@vocab (long list). Long list and _other cannot be combined.
            </assert>
        </rule>
    </pattern>
</schema>
