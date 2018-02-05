// Sometimes, you just want to hack some JS.

window.onload = function() {
  var codeElements = document.getElementsByTagName('code')
  for (var i = 0; i < codeElements.length; ++i) {
    if (codeElements[i].textContent.startsWith('---')) {
      process(codeElements[i])
    }
  }
}

function h3 (content) {
  var h3Node = document.createElement('h3')
  h3Node.textContent = content
  return h3Node
}

function textarea (content) {
  var textareaNode = document.createElement('textarea')
  textareaNode.textContent = content
  textareaNode.setAttribute('rows', 15)
  textareaNode.setAttribute('cols', 80)
  return textareaNode
}

function process (originalNode) {
  var div = originalNode.parentNode
  div.removeChild(originalNode)

  var form = div.appendChild(document.createElement('form'))

  var yamlSchemaNode = form
    .appendChild(document.createElement('div'))
    .appendChild(textarea(originalNode.textContent))

  var jsonSchemaNode = form
    .appendChild(document.createElement('div'))
    .appendChild(textarea('foo'))
  jsonSchemaNode.setAttribute('readonly', true)

  var buttonDiv = form.appendChild(document.createElement('div'))
  var showJsonSchema = buttonDiv.appendChild(document.createElement('button'))
  showJsonSchema.setAttribute('type', 'button')
  showJsonSchema.textContent = 'Show generated JSON schema'
  showJsonSchema.addEventListener('click', function () {
    showYamlSchema.style.display = ''
    showJsonSchema.style.display = 'none'
    yamlSchemaNode.style.display = 'none'
    jsonSchemaNode.style.display = ''
    return false
  })
  var showYamlSchema = buttonDiv.appendChild(document.createElement('button'))
  showYamlSchema.setAttribute('type', 'button')
  showYamlSchema.textContent = 'Show original yajsonschema input'
  showYamlSchema.addEventListener('click', function () {
    showYamlSchema.style.display = 'none'
    showJsonSchema.style.display = ''
    yamlSchemaNode.style.display = ''
    jsonSchemaNode.style.display = 'none'
    return false
  })
  showYamlSchema.dispatchEvent(new Event('click'))

  // Spacing hack... (don't want to understand default template CSS)
  form.appendChild(document.createElement('p'))

  var validationNode = form
    .appendChild(document.createElement('div'))
  validationNode.appendChild(h3('JSON to validate'))
  var jsonToValidate = validationNode.appendChild(textarea('{}'))
  validationNode.appendChild(h3('Result'))
  var result = validationNode.appendChild(document.createElement('code'))

  setUpChangeListeners(yamlSchemaNode, jsonSchemaNode, jsonToValidate, result)
}

function setUpChangeListeners (yamlSchemaNode, jsonSchemaNode, jsonToValidate, result) {
  yamlSchemaNode.addEventListener('input', function () {
    jsonSchemaNode.value = yajsonschema.convert(yamlSchemaNode.value)
  })
  yamlSchemaNode.dispatchEvent(new Event('input'))

  jsonToValidate.addEventListener('input', function () {
    result.textContent = yajsonschema.validate(yamlSchemaNode.value, jsonToValidate.value)
  })
  jsonToValidate.dispatchEvent(new Event('input'))
}
